package user

import (
	"github.com/markbates/goth"
	logging "github.com/op/go-logging"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/exceptions"
	"github.com/tryanzu/core/modules/helpers"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"errors"
	"regexp"
	"strings"
	"time"
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

	var model *UserPrivate
	context := module
	database := deps.Container.Mgo()

	switch usr.(type) {
	case bson.ObjectId:

		// Get the user using it's id
		err := database.C("users").FindId(usr.(bson.ObjectId)).One(&model)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid user id. Not found."}
		}

	case bson.M:

		// Get the user using it's id
		err := database.C("users").Find(usr.(bson.M)).One(&model)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid user id. Not found."}
		}

	case *UserPrivate:

		model = usr.(*UserPrivate)

	default:
		panic("Unkown argument")
	}

	user := &One{data: model, di: context}
	return user, nil
}

type Store interface {
	Insert(docs ...interface{}) error
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
	hashed, err := helpers.HashPassword(password)
	if err != nil {
		return nil, err
	}
	usr := &UserPrivate{
		User: User{
			Id:          bson.NewObjectId(),
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
	err = dal.Insert(usr)
	if err != nil {
		return nil, err
	}
	return usr, nil
}

// SignUp a user with email and username checks
func (module *Module) SignUp(email, username, password, referral string) (*One, error) {
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
	unique, err := deps.Container.Mgo().C("users").Find(bson.M{
		"$or": []bson.M{
			{"email": email},
			{"username": bson.RegEx{
				Pattern: regexp.QuoteMeta(username),
				Options: "i",
			}},
		},
	}).Count()
	if unique > 0 || err != nil {
		return nil, exceptions.OutOfBounds{
			Msg: "User already exists.",
		}
	}
	usr, err := InsertUser(deps.Container.Mgo().C("users"), username, password, email)
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
	id := bson.NewObjectId()

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

	err = deps.Container.Mgo().C("users").Insert(usr)
	if err != nil {
		panic(err)
	}

	err = deps.Container.Mgo().C("users").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{provider: user.RawData}})
	if err != nil {
		panic(err)
	}

	_user := &One{data: usr, di: module}

	return _user, nil
}

func (module *Module) IsValidRecoveryToken(token string) (bool, error) {
	// Only tokens 15 minutes old are valid
	c, err := deps.Container.Mgo().C("user_recovery_tokens").Find(
		bson.M{
			"token":      token,
			"used":       false,
			"created_at": bson.M{"$gte": time.Now().Add(-15 * time.Minute)},
		},
	).Count()
	return c > 0, err
}

func (module *Module) GetUserFromRecoveryToken(token string) (*One, error) {

	var model UserRecoveryToken

	database := deps.Container.Mgo()
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"used": true, "updated_at": time.Now()}},
		ReturnNew: false,
	}

	_, err := database.C("user_recovery_tokens").Find(bson.M{"token": token}).Apply(change, &model)

	if err != nil {
		return nil, err
	}

	usr, err := module.Get(model.UserId)

	if err != nil {
		return nil, err
	}

	return usr, nil
}
