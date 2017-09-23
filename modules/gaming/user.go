package gaming

import (
	notify "github.com/fernandez14/spartangeek-blacker/board/notifications"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"gopkg.in/mgo.v2/bson"
	"time"
)

// Compute rewards for user making one pos.
func UserHasPublished(d Deps, id bson.ObjectId) (err error) {
	return IncreaseUserSwords(d, id, 1)
}

// Compute rewards for user making one comment.
func UserHasCommented(d Deps, id bson.ObjectId) error {
	return IncreaseUserSwords(d, id, 0)
}

// Increase swords of given user (id).
func IncreaseUserSwords(d Deps, id bson.ObjectId, swords int) (err error) {
	if swords > 0 {
		users := d.Mgo().C("users")

		// Perform update using $inc operator.
		err = users.Update(bson.M{"_id": id}, bson.M{"$inc": bson.M{"gaming.swords": swords}})
		if err != nil {
			return
		}
	}

	err = syncLevelStats(d, id, false)
	return
}

// Reset gamification temporal stuff based on user level.
func syncLevelStats(d Deps, id bson.ObjectId, forceSync bool) (err error) {
	var usr struct {
		G user.UserGaming `bson:"gaming"`
	}

	users := d.Mgo().C("users")
	fields := bson.M{"gaming.swords": 1, "gaming.level": 1, "gaming.tribute": 1}
	err = users.FindId(id).Select(fields).One(&usr)
	if err != nil {
		return
	}

	swords := usr.G.Swords
	level := usr.G.Level
	for _, r := range d.GamingConfig().Rules {

		// Continue when out of bounds.
		if swords > r.End || swords < r.Start {
			continue
		}

		// The user level just changed
		if level != r.Level || forceSync {
			update := bson.M{
				"gaming.level":   r.Level,
				"gaming.shit":    r.Shit,
				"gaming.tribute": r.Tribute,
			}

			// Update the user gamification facts
			err = users.Update(bson.M{"_id": id}, bson.M{"$set": update})
			if err != nil {
				return
			}

			// Send updated data over the wire
			notify.Transmit <- notify.Socket{"user " + id.Hex(), "gaming", map[string]interface{}{
				"level":   r.Level,
				"tribute": r.Tribute,
				"shit":    r.Shit,
			}}
		}

		break
	}

	return
}

type User struct {
	user *user.One
	di   *Module
}

// Sync user gamification relevant facts
func (self *User) SyncToLevel(reset bool) {

	database := self.di.Mongo.Database
	rules := self.di.Rules.Rules
	user := self.user.Data()

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

func (self *User) Sync() {

	defer self.di.Errors.Recover()

	id := self.user.Data().Id.Hex()
	validated := self.user.Data().Validated

	// Sync user stuff
	self.di.Firebase.Set("users/"+id+"/validated", validated, nil)
}
