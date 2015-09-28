package cli

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"gopkg.in/mgo.v2/bson"
	"log"
	"fmt"
	"regexp"
)

type Module struct {
	Mongo  *mongo.Service               `inject:""`
	Errors *exceptions.ExceptionsModule `inject:""`
}

type fn func()

func (module Module) Run(name string) {

	commands := map[string]fn {
		"slug-fix": module.SlugFix,
		"codes-fix": module.Codes,
	}

	if handler, exists := commands[name]; exists {

		handler()
		return
	} 

	// If reachs this point then panic
	log.Panic("No such handler for cli")
}

func (module Module) SlugFix() {

	var usr user.User
	database := module.Mongo.Database
	valid_name, _ := regexp.Compile(`^[a-zA-Z][a-zA-Z0-9]*[._-]?[a-zA-Z0-9]+$`)

	// Get all users
	iter := database.C("users").Find(nil).Select(bson.M{"_id": 1, "username": 1, "email": 1, "username_slug": 1}).Iter()

	for iter.Next(&usr) {

		slug := helpers.StrSlug(usr.UserName)

		if !valid_name.MatchString(usr.UserName) {

			// Fallback username to slug
			err := database.C("users").Update(bson.M{"_id": usr.Id}, bson.M{"$set": bson.M{"username": slug, "username_slug": slug, "name_changes": 0}})

			if err != nil {
				panic(err)
			}

				
			log.Printf("\n%v --> %v\n", usr.UserName, slug)		
			continue

		} else {

			// Fix slug in case they need if
			if slug != usr.UserNameSlug {
fmt.Printf("+")
				err := database.C("users").Update(bson.M{"_id": usr.Id}, bson.M{"$set": bson.M{"username_slug": slug}})

				if err != nil {
					panic(err)
				}

				fmt.Printf("-")
				log.Printf("\n%v -slug-> %v\n", usr.UserNameSlug, slug)		
				continue
			}
		}

		fmt.Printf(".")
	}
}


func (module Module) Codes() {

	var usr user.User
	database := module.Mongo.Database

	// Get all users
	iter := database.C("users").Find(nil).Select(bson.M{"_id": 1, "ref_code": 1, "ver_code": 1}).Iter()

	for iter.Next(&usr) {

		if usr.VerificationCode == "" {

			code := helpers.StrRandom(12)
			err := database.C("users").Update(bson.M{"_id": usr.Id}, bson.M{"$set": bson.M{"ver_code": code, "validated": false}})

			if err != nil {
				panic(err)
			}

			fmt.Printf("+")
		}

		if usr.ReferralCode == "" {

			code := helpers.StrRandom(6)
			err := database.C("users").Update(bson.M{"_id": usr.Id}, bson.M{"$set": bson.M{"ref_code": code}})

			if err != nil {
				panic(err)
			}

			fmt.Printf("-")
		}

		if usr.ReferralCode != "" && usr.VerificationCode != "" {

			fmt.Printf(".")
		}
	}
}