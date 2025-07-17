package user

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/markbates/goth"
	logging "github.com/op/go-logging"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/exceptions"
	"github.com/tryanzu/core/modules/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Boot() *Module {
	return &Module{}
}

type Module struct {
	Errors *exceptions.ExceptionsModule `inject:""`
	Logger *logging.Logger              `inject:""`
}

var (
	validUsername = regexp.MustCompile(`^[a-zA-Z]+([_.-]?[a-zA-Z0-9])*$`)
)

// Gets an instance of a user
func (module *Module) Get(usr interface{}) (*One, error) {
	ctx := context.Background()
	var model *UserPrivate
	ctx_module := module
	database := deps.Container.Mgo()
	collection := database.Collection("users")

	switch usr := usr.(type) {
	case primitive.ObjectID:
		// Get the user using its id
		err := collection.FindOne(ctx, bson.M{"_id": usr}).Decode(&model)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, exceptions.NotFound{Msg: "Invalid user id. Not found."}
			}
			return nil, err
		}

	case bson.M:
		// Get the user using the filter
		err := collection.FindOne(ctx, usr).Decode(&model)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, exceptions.NotFound{Msg: "Invalid user id. Not found."}
			}
			return nil, err
		}

	case *UserPrivate:
		model = usr

	default:
		panic("Unknown argument")
	}

	user := &One{data: model, di: ctx_module}
	return user, nil
}

type Store interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
}

type Opt func(*UserPrivate)

func Validated(v bool) Opt {
	return func(up *UserPrivate) {
		up.Validated = v
	}
}

func WithRole(role string) Opt {
	return func(up *UserPrivate) {
		var f bool
		for _, v := range up.Roles {
			if v.Name == role {
				f = true
				break
			}
		}
		if !f {
			up.Roles = append(up.Roles, UserRole{
				Name: role,
			})
		}
	}
}

// InsertUser creates a new user with the provided username, password, and email.
// It generates a new user ID, sets the initial user properties, and inserts the
// user into the database. If the insert fails, it returns an error.
func InsertUser(dal Store, username, password, email string, opts ...Opt) (*UserPrivate, error) {
	ctx := context.Background()
	hashed, err := helpers.HashPassword(password)
	if err != nil {
		return nil, err
	}
	usr := &UserPrivate{
		User: User{
			Id:          primitive.NewObjectID(),
			UserName:    username,
			Description: "",
			Profile: map[string]interface{}{
				"country": "",
				"bio":     "",
			},
			Created:     time.Now(),
			Permissions: make([]string, 0),
			NameChanges: 1,
			Roles: []UserRole{
				{
					Name: "user",
				},
			},
			Gaming: UserGaming{
				Swords: 1,
			},
			Validated: false,
		},
		Password:           hashed,
		Email:              email,
		EmailNotifications: true,
		ReferralCode:       helpers.StrRandom(6),
		VerificationCode:   helpers.StrRandom(12),
		Updated:            time.Now(),
	}
	for _, fn := range opts {
		fn(usr)
	}
	_, err = dal.InsertOne(ctx, usr)
	if err != nil {
		return nil, err
	}
	return usr, nil
}

// SignUp a user with email and username checks
func (module *Module) SignUp(email, username, password, referral string) (*One, error) {
	ctx := context.Background()
	if !validUsername.MatchString(username) || strings.Count(username, "") < 3 || strings.Count(username, "") > 21 {
		return nil, exceptions.OutOfBounds{
			Msg: "Invalid username. Must have only alphanumeric characters.",
		}
	}
	if !helpers.IsEmail(email) {
		return nil, exceptions.OutOfBounds{
			Msg: "Invalid email. Provide a real one.",
		}
	}

	// Check if user already exists using that email
	database := deps.Container.Mgo()
	collection := database.Collection("users")
	filter := bson.M{
		"$or": []bson.M{
			{"email": email},
			{"username": bson.M{
				"$regex":   regexp.QuoteMeta(username),
				"$options": "i",
			}},
		},
	}
	unique, err := collection.CountDocuments(ctx, filter)
	if unique > 0 || err != nil {
		return nil, exceptions.OutOfBounds{
			Msg: "User already exists.",
		}
	}
	usr, err := InsertUser(collection, username, password, email)
	if err != nil {
		panic(err)
	}

	user := &One{data: usr, di: module}

	return user, nil
}

func (m *Module) computeNickname(nicknames ...string) (string, error) {
	var nickname string
	for _, name := range nicknames {
		if len(nickname) > 0 {
			break
		}

		if len(strings.TrimSpace(name)) > 0 {
			nickname = strings.TrimSpace(name)
		}
	}

	nickname = helpers.StrSlug(nickname)
	if len(nickname) == 0 {
		return "", errors.New("Could not compute nickname from empty strings")
	}

	return nickname, nil
}

// Sign up user from oauth provider
func (module *Module) OauthSignup(provider string, user goth.User) (*One, error) {
	ctx := context.Background()
	id := primitive.NewObjectID()

	profile := map[string]interface{}{
		"country": "",
		"bio":     "",
	}

	nickname, err := module.computeNickname(user.NickName, user.Name, "User"+helpers.StrRandom(8))
	if err != nil {
		return nil, err
	}

	usr := &UserPrivate{
		User: User{
			Id:          id,
			UserName:    nickname,
			Description: "",
			Profile:     profile,
			Created:     time.Now(),
			Permissions: make([]string, 0),
			NameChanges: 0,
			Roles: []UserRole{
				{
					Name: "user",
				},
			},
			Validated: true,
		},
		Password:         "",
		Email:            user.Email,
		ReferralCode:     helpers.StrRandom(6),
		VerificationCode: helpers.StrRandom(12),
		Updated:          time.Now(),
	}
	if len(usr.Email) > 0 {
		usr.EmailNotifications = true
	}

	database := deps.Container.Mgo()
	collection := database.Collection("users")
	_, err = collection.InsertOne(ctx, usr)
	if err != nil {
		panic(err)
	}

	update := bson.M{"$set": bson.M{provider: user.RawData}}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		panic(err)
	}

	_user := &One{data: usr, di: module}

	return _user, nil
}

func (module *Module) IsValidRecoveryToken(token string) (bool, error) {
	ctx := context.Background()
	// Only tokens 15 minutes old are valid
	database := deps.Container.Mgo()
	collection := database.Collection("user_recovery_tokens")
	filter := bson.M{
		"token":      token,
		"used":       false,
		"created_at": bson.M{"$gte": time.Now().Add(-15 * time.Minute)},
	}
	c, err := collection.CountDocuments(ctx, filter)
	return c > 0, err
}

func (module *Module) GetUserFromRecoveryToken(token string) (*One, error) {
	ctx := context.Background()
	var model UserRecoveryToken

	database := deps.Container.Mgo()
	collection := database.Collection("user_recovery_tokens")
	filter := bson.M{"token": token}
	update := bson.M{"$set": bson.M{"used": true, "updated_at": time.Now()}}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.Before)
	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&model)
	if err != nil {
		return nil, err
	}

	usr, err := module.Get(model.UserId)
	if err != nil {
		return nil, err
	}

	return usr, nil
}
