package user

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"gopkg.in/mgo.v2/bson"
	"regexp"
	"strings"
	"time"
)

func Boot() *Module {

	module := &Module{}

	return module
}

type Module struct {
	Mongo *mongo.Service `inject:""`
	Mail  *mail.Module   `inject:""`
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
			Id:               id,
			UserName:         username,
			UserNameSlug:     slug,
			Description:      "",
			Profile:          profile,
			Created:          time.Now(),
			Permissions:      make([]string, 0),
			NameChanges:      1,
			Roles: []UserRole{
				{
					Name: "user",
				},
			},
			Validated:        false,
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

func (module *Module) SignUpFacebook(facebook map[string]interface{}) (*One, error) {

	context := module
	database := module.Mongo.Database
	id := bson.NewObjectId()

	// Track the referral if we have to
	if _, has_referral := facebook["ref"]; has_referral {

		var reference User

		referral := facebook["ref"].(string)
		err := database.C("users").Find(bson.M{"ref_code": referral}).One(&reference)

		// Track the referral link
		if err == nil {

			track := &ReferralModel{
				OwnerId:   reference.Id,
				UserId:    id,
				Code:      referral,
				Confirmed: true,
				Created:   time.Now(),
				Updated:   time.Now(),
			}

			err := database.C("referrals").Insert(track)

			if err != nil {

				panic(err)
			}
		}
	}

	email := ""

	if _, ok := facebook["email"]; ok {

		email = facebook["email"].(string)
	}

	profile := map[string]interface{}{
		"country": "México",
		"bio":     "Just another spartan geek",
	}

	username := facebook["first_name"].(string) + " " + facebook["last_name"].(string)
	username = helpers.StrSlug(username)

	usr := &UserPrivate{
		User: User{
			Id:               id,
			UserName:         username,
			UserNameSlug:     username,
			Description:      "",
			Profile:          profile,
			Created:          time.Now(),
			Permissions:      make([]string, 0),
			NameChanges:      0,
			Roles: []UserRole{
				{
					Name: "user",
				},
			},
			Validated:        true,
		},
		Password:         "",
		Email:            email,
		ReferralCode:     helpers.StrRandom(6),
		VerificationCode: helpers.StrRandom(12),
		Updated:          time.Now(),
		Facebook:         facebook,
	}

	err := database.C("users").Insert(usr)

	if err != nil {
		panic(err)
	}

	user := &One{data: usr, di: context}

	return user, nil
}
