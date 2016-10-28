package user

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/payments"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/markbates/goth"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/op/go-logging.v1"

	"errors"
	"regexp"
	"strings"
	"time"
)

func Boot() *Module {

	module := &Module{}

	return module
}

type Module struct {
	Mongo  *mongo.Service               `inject:""`
	Mail   *mail.Module                 `inject:""`
	Errors *exceptions.ExceptionsModule `inject:""`
	Logger *logging.Logger              `inject:""`
}

// Gets an instance of a user
func (module *Module) Get(usr interface{}) (*One, error) {

	var model *UserPrivate
	context := module
	database := module.Mongo.Database

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

func (module *Module) GetTopDonators() []map[string]interface{} {

	var payments []payments.Payment

	database := module.Mongo.Database
	err := database.C("payments").Find(bson.M{"type": "donation", "status": "confirmed"}).Sort("-amount").Limit(20).All(&payments)

	if err != nil {
		panic(err)
	}

	var list []bson.ObjectId

	for _, p := range payments {
		list = append(list, p.UserId)
	}

	var users []UserSimple

	usersMap := map[string]UserSimple{}
	err = database.C("users").Find(bson.M{"_id": bson.M{"$in": list}}).Select(UserSimpleFields).All(&users)

	if err != nil {
		panic(err)
	}

	for _, u := range users {
		usersMap[u.Id.Hex()] = u
	}

	data := []map[string]interface{}{}

	for _, p := range payments {

		usr := usersMap[p.UserId.Hex()]
		data = append(data, map[string]interface{}{
			"created_at": p.Created,
			"amount":     p.Amount,
			"user":       usr,
		})
	}

	return data
}

// Signup a user with email and username checks
func (module *Module) SignUp(email, username, password, referral string) (*One, error) {

	context := module
	database := module.Mongo.Database
	slug := helpers.StrSlug(username)
	valid_name, err := regexp.Compile(`^[a-zA-Z][a-zA-Z0-9]*[._-]?[a-zA-Z0-9]+$`)

	if err != nil {
		panic(err)
	}

	hash := helpers.Sha256(password)
	id := bson.NewObjectId()

	if valid_name.MatchString(username) == false || strings.Count(username, "") < 3 || strings.Count(username, "") > 21 {
		return nil, exceptions.OutOfBounds{"Invalid username. Must have only alphanumeric characters."}
	}

	if helpers.IsEmail(email) == false {
		return nil, exceptions.OutOfBounds{"Invalid email. Provide a real one."}
	}

	// Check if user already exists using that email
	unique, _ := database.C("users").Find(bson.M{"$or": []bson.M{{"email": email}, {"username_slug": slug}}}).Count()

	if unique > 0 {
		return nil, exceptions.OutOfBounds{"User already exists."}
	}

	// Track the referral if we have to
	if referral != "" {
		var reference User
		err := database.C("users").Find(bson.M{"ref_code": referral}).One(&reference)

		// Track the referral link
		if err == nil {
			track := &ReferralModel{
				OwnerId:   reference.Id,
				UserId:    id,
				Code:      referral,
				Confirmed: false,
				Created:   time.Now(),
				Updated:   time.Now(),
			}

			err := database.C("referrals").Insert(track)

			if err != nil {
				panic(err)
			}
		}
	}

	profile := map[string]interface{}{
		"country": "México",
		"bio":     "Just another spartan geek",
	}

	usr := &UserPrivate{
		User: User{
			Id:           id,
			UserName:     username,
			UserNameSlug: slug,
			Description:  "",
			Profile:      profile,
			Created:      time.Now(),
			Permissions:  make([]string, 0),
			NameChanges:  1,
			Roles: []UserRole{
				{
					Name: "user",
				},
			},
			Validated: false,
		},
		Password:         hash,
		Email:            email,
		ReferralCode:     helpers.StrRandom(6),
		VerificationCode: helpers.StrRandom(12),
		Updated:          time.Now(),
	}

	err = database.C("users").Insert(usr)

	if err != nil {
		panic(err)
	}

	user := &One{data: usr, di: context}

	// Send the confirmation email in other thread
	go user.SendConfirmationEmail()

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
	db := module.Mongo.Database
	id := bson.NewObjectId()

	profile := map[string]interface{}{
		"country": "México",
		"bio":     "Just another spartan geek",
	}

	nickname, err := module.computeNickname(user.NickName, user.Name, "User"+helpers.StrRandom(8))
	if err != nil {
		return nil, err
	}

	usr := &UserPrivate{
		User: User{
			Id:           id,
			UserName:     nickname,
			UserNameSlug: nickname,
			Description:  "",
			Profile:      profile,
			Created:      time.Now(),
			Permissions:  make([]string, 0),
			NameChanges:  0,
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

	err = db.C("users").Insert(usr)
	if err != nil {
		panic(err)
	}

	err = db.C("users").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{provider: user.RawData}})
	if err != nil {
		panic(err)
	}

	_user := &One{data: usr, di: module}

	return _user, nil
}

func (module *Module) IsValidRecoveryToken(token string) (bool, error) {

	database := module.Mongo.Database

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

	database := module.Mongo.Database
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
