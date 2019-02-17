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

// SignUp a user with email and username checks
func (module *Module) SignUp(email, username, password, referral string) (*One, error) {
	id := bson.NewObjectId()
	if validUsername.MatchString(username) == false || strings.Count(username, "") < 3 || strings.Count(username, "") > 21 {
		return nil, exceptions.OutOfBounds{
			Msg: "Invalid username. Must have only alphanumeric characters.",
		}
	}

	if helpers.IsEmail(email) == false {
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

	hashed, err := helpers.HashPassword(password)
	if err != nil {
		return nil, err
	}

	profile := map[string]interface{}{
		"country": "",
		"bio":     "",
	}

	usr := &UserPrivate{
		User: User{
			Id:          id,
			UserName:    username,
			Description: "",
			Profile:     profile,
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
		Password:         hashed,
		Email:            email,
		ReferralCode:     helpers.StrRandom(6),
		VerificationCode: helpers.StrRandom(12),
		Updated:          time.Now(),
	}

	err = deps.Container.Mgo().C("users").Insert(usr)
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

	database := deps.Container.Mgo()

	// Only tokens that are 15 minutes old
	valid_date := time.Now().Add(-15 * time.Minute)

	c, err := database.C("user_recovery_tokens").Find(bson.M{"token": token, "used": false, "created_at": bson.M{"$gte": valid_date}}).Count()

	if err != nil {
		return false, err
	}

	return c > 0, nil
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
