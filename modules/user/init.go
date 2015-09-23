package user

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"gopkg.in/mgo.v2/bson"
	"regexp"
	"time"
)

func Boot() *Module {

	module := &Module{}

	return module
}

type Module struct {
	Mongo *mongo.Service `inject:""`
}

// Gets an instance of a user
func (module *Module) Get(id bson.ObjectId) (*One, error) {

	var model *User
	context := module
	database := module.Mongo.Database

	// Get the user using it's id
	err := database.C("users").FindId(id).One(&model)

	if err != nil {

		return nil, exceptions.NotFound{"Invalid user id. Not found."}
	}

	user := &One{data: model, di: context}

	return user, nil
}

// Signup a user with email and username checks
func (module *Module) SignUp(email, username, password, referral string) (*One, error) {

	context := module
	database := module.Mongo.Database
	slug := helpers.StrSlug(username)
	valid_name, _ := regexp.Compile(`^[0-9a-zA-Z\-]{0,32}$`)
	hash := helpers.Sha256(password)
	id := bson.NewObjectId()

	// Check if user already exists using that email
	unique, _ := database.C("users").Find(bson.M{"$or": []bson.M{{"email": email}, {"username_slug": slug}}}).Count()

	if unique > 0 {

		return nil, exceptions.OutOfBounds{"User already exists."}
	}

	if valid_name.MatchString(username) == false {

		return nil, exceptions.OutOfBounds{"Invalid username. Must have only alphanumeric characters."}
	}

	// Track the referral if we have to
	if referral != "" {

		var reference User

		err := database.C("users").Find(bson.M{"ref_code": referral}).One(&reference)

		// Track the referral link
		if err == nil {

			track := &ReferralModel{
				OwnerId: reference.Id,
				UserId:  id,
				Code:    referral,
				Created: time.Now(),
				Updated: time.Now(),
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

	usr := &User{
		Id:               id,
		UserName:         username,
		UserNameSlug:     slug,
		NameChanges:      1,
		Password:     	  hash,
		Email:            email,
		Permissions:      make([]string, 0),
		Description:      "",
		Profile:          profile,
		ReferralCode:     helpers.StrRandom(6),
		VerificationCode: helpers.StrRandom(12),
		Validated: 	      false,
		Created:          time.Now(),
		Updated:          time.Now(),
		Roles: []UserRole{
			{
				Name: "user",
			},
		},
	}

	err := database.C("users").Insert(usr)

	if err != nil {
		panic(err)
	}

	user := &One{data: usr, di: context}

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
				OwnerId: reference.Id,
				UserId:  id,
				Code:    referral,
				Created: time.Now(),
				Updated: time.Now(),
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

	usr := &User{
		Id:               id,
		UserName:         username,
		UserNameSlug:     username,
		NameChanges:      0,
		Password:         "",
		Email:            email,
		Permissions:      make([]string, 0),
		Description:      "",
		Profile:          profile,
		Facebook:         facebook,
		ReferralCode:     helpers.StrRandom(6),
		VerificationCode: helpers.StrRandom(12),
		Validated: 	      true,
		Created:      time.Now(),
		Updated:      time.Now(),
		Roles: []UserRole{
			{
				Name: "user",
			},
		},
	}

	err := database.C("users").Insert(usr)

	if err != nil {
		panic(err)
	}

	user := &One{data: usr, di: context}

	return user, nil
}
