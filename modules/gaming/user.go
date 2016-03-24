package gaming

import (
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type User struct {
	user *user.One
	di   *Module
}

// Update user gamification stats because of something he did
func (self *User) Did(action string) {

	defer self.di.Errors.Recover()

	var swords int

	switch action {
	case "comment":
		swords = 1
	case "publish":
		swords = 1
	}

	self.Swords(swords)
}

// Sync user gamification relevant facts
func (self *User) SyncToLevel(reset bool) {

	database := self.di.Mongo.Database
	rules := self.di.Rules.Rules
	user := self.user.Data()
	activated := user.Validated
	user.Gaming.Active = activated

	// Swords and level will determine current user level
	user_swords := user.Gaming.Swords
	user_level := user.Gaming.Level

	for _, rule := range rules {

		// Determine the user level by his swords
		if rule.Start <= user_swords && rule.End >= user_swords {

			// The user level just changed
			if user_level != rule.Level || reset {

				fact_set := make(bson.M)

				fact_set["gaming.level"] = rule.Level
				user.Gaming.Level = rule.Level

				// Only reset when truly needed
				if rule.Shit >= user.Gaming.Shit {

					fact_set["gaming.shit"] = rule.Shit
					user.Gaming.Shit = rule.Shit
				}

				if rule.Tribute >= user.Gaming.Tribute {

					fact_set["gaming.tribute"] = rule.Tribute
					user.Gaming.Tribute = rule.Tribute
				}

				// Update the user gamification facts
				err := database.C("users").Update(bson.M{"_id": user.Id}, bson.M{"$set": fact_set})

 				if err != nil {
 					panic(err)
 				}

				// Runtime update
				self.user.RUpdate(user)
			}

			break
		}
	}

	self.SyncRealtimeFirebase(user.Gaming)
}

// Does the daily login logic for the user
func (self *User) DailyLogin() {

	database := self.di.Mongo.Database
	rules := self.di.Rules.Rules
	usr := self.user.Data()

	// Dates and stuff
	duration := time.Since(usr.Gamificated)

	if duration.Hours() >= 24 {

		for _, rule := range rules {

			if rule.Level == usr.Gaming.Level {

				self.Coins(rule.Coins)
				self.SyncToLevel(true)
				break
			}
		}

		// Update gamificated at
		err := database.C("users").Update(bson.M{"_id": usr.Id}, bson.M{"$set": bson.M{"gamificated_at": time.Now()}})

		if err != nil {
			panic(err)
		}
	}
}

// Increases or decreases user swords
func (self *User) Swords(how_many int) {

	// Recover from any panic even inside this goroutine
	defer self.di.Errors.Recover()

	// Get the database interface from the DI
	database := self.di.Mongo.Database
	err := database.C("users").Update(bson.M{"_id": self.user.Data().Id}, bson.M{"$inc": bson.M{"gaming.swords": how_many}})

	if err != nil {
		panic(err)
	}

	// Runtime update
	user := self.user.Data()

	user.Gaming.Swords = user.Gaming.Swords + how_many
	self.user.RUpdate(user)

	// Check for level changes and stuff
	self.SyncToLevel(false)
}

// Increases or decreases user coins
func (self *User) Coins(how_many int) {

	// Recover from any panic even inside this goroutine
	defer self.di.Errors.Recover()

	// Get the database interface from the DI
	database := self.di.Mongo.Database

	err := database.C("users").Update(bson.M{"_id": self.user.Data().Id}, bson.M{"$inc": bson.M{"gaming.coins": how_many}})

	if err != nil {
		panic(err)
	}

	// Runtime update
	user := self.user.Data()
	user.Gaming.Coins = user.Gaming.Coins + how_many

	self.user.RUpdate(user)

	// Check for level changes and stuff
	self.SyncToLevel(false)
}

// Increases or decreases user tribute
func (self *User) Tribute(how_many int) {

	// Recover from any panic even inside this goroutine
	defer self.di.Errors.Recover()

	// Get the database interface from the DI
	database := self.di.Mongo.Database
	err := database.C("users").Update(bson.M{"_id": self.user.Data().Id}, bson.M{"$inc": bson.M{"gaming.tribute": how_many}})

	if err != nil {
		panic(err)
	}

	// Runtime update
	user := self.user.Data()
	user.Gaming.Tribute = user.Gaming.Tribute + how_many
	self.user.RUpdate(user)

	// Check for level changes and stuff
	self.SyncToLevel(false)
}

// Increases or decreases user shit
func (self *User) Shit(how_many int) {

	// Recover from any panic even inside this goroutine
	defer self.di.Errors.Recover()

	// Get the database interface from the DI
	database := self.di.Mongo.Database
	err := database.C("users").Update(bson.M{"_id": self.user.Data().Id}, bson.M{"$inc": bson.M{"gaming.shit": how_many}})

	if err != nil {
		panic(err)
	}

	// Runtime update
	user := self.user.Data()
	user.Gaming.Shit = user.Gaming.Shit + how_many
	self.user.RUpdate(user)

	// Check for level changes and stuff
	self.SyncToLevel(false)
}

func (self *User) SyncRealtimeFirebase(data user.UserGaming) {

	// Recover from any panic even inside this goroutine
	defer self.di.Errors.Recover()

	// Get the user path from firebase
	userPath := "users/" + self.user.Data().Id.Hex()

	// Update the gaming part
	self.di.Firebase.Set(userPath+"/gaming", data, nil)
}
