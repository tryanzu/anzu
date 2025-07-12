package gaming

import (
	"context"
	"time"

	notify "github.com/tryanzu/core/board/notifications"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IncreaseUserSwords for given id.
func IncreaseUserSwords(d Deps, id primitive.ObjectID, swords int) error {
	return increaseUserAttr(d, id, "gaming.swords", swords)
}

// IncreaseUserCoins for given id.
func IncreaseUserCoins(d Deps, id primitive.ObjectID, coins int) error {
	return increaseUserAttr(d, id, "gaming.coins", coins)
}

// IncreaseUserTribute for given id.
func IncreaseUserTribute(d Deps, id primitive.ObjectID, tribute int) error {
	return increaseUserAttr(d, id, "gaming.tribute", tribute)
}

func increaseUserAttr(d Deps, id primitive.ObjectID, field string, n int) (err error) {
	if n == 0 {
		// ignore.
		return nil
	}
	ctx := context.Background()
	// Perform update using $inc operator.
	_, err = d.Mgo().Collection("users").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{field: n}})
	if err != nil {
		return
	}

	// Fix reputation attr when is less than ($lt) 0
	_, err = d.Mgo().Collection("users").UpdateOne(ctx, bson.M{"_id": id, field: bson.M{"$lt": 0}}, bson.M{"$set": bson.M{field: 0}})
	if err != nil {
		return
	}
	err = syncLevelStats(d, id, false)
	return
}

// Reset gamification temporal stuff based on user level.
func syncLevelStats(d Deps, id primitive.ObjectID, forceSync bool) (err error) {
	var usr struct {
		G user.UserGaming `bson:"gaming"`
	}

	ctx := context.Background()
	users := d.Mgo().Collection("users")
	fields := bson.M{"gaming.swords": 1, "gaming.level": 1, "gaming.tribute": 1}
	err = users.FindOne(ctx, bson.M{"_id": id}, options.FindOne().SetProjection(fields)).Decode(&usr)
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
			_, err = users.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
			if err != nil {
				return
			}

			// Send updated data over the wire
			notify.Transmit <- notify.Socket{"user " + id.Hex(), "gaming", map[string]interface{}{
				"level":   r.Level,
				"tribute": r.Tribute,
				"shit":    r.Shit,
				"swords":  swords,
			}}
			break
		}

		notify.Transmit <- notify.Socket{"user " + id.Hex(), "gaming", map[string]interface{}{
			"level":   usr.G.Level,
			"tribute": usr.G.Tribute,
			"shit":    usr.G.Shit,
			"swords":  usr.G.Swords,
		}}

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

	ctx := context.Background()
	database := deps.Container.Mgo()
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
				_, err := database.Collection("users").UpdateOne(ctx, bson.M{"_id": user.Id}, bson.M{"$set": fact_set})

				if err != nil {
					panic(err)
				}

				// Runtime update
				self.user.RUpdate(user)
			}

			break
		}
	}
}

// Does the daily login logic for the user
func (self *User) DailyLogin() {

	ctx := context.Background()
	database := deps.Container.Mgo()
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
		_, err := database.Collection("users").UpdateOne(ctx, bson.M{"_id": usr.Id}, bson.M{"$set": bson.M{"gamificated_at": time.Now()}})

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
	ctx := context.Background()
	database := deps.Container.Mgo()
	_, err := database.Collection("users").UpdateOne(ctx, bson.M{"_id": self.user.Data().Id}, bson.M{"$inc": bson.M{"gaming.swords": how_many}})

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
	ctx := context.Background()
	database := deps.Container.Mgo()

	_, err := database.Collection("users").UpdateOne(ctx, bson.M{"_id": self.user.Data().Id}, bson.M{"$inc": bson.M{"gaming.coins": how_many}})

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
	ctx := context.Background()
	database := deps.Container.Mgo()
	_, err := database.Collection("users").UpdateOne(ctx, bson.M{"_id": self.user.Data().Id}, bson.M{"$inc": bson.M{"gaming.tribute": how_many}})

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
	ctx := context.Background()
	database := deps.Container.Mgo()
	_, err := database.Collection("users").UpdateOne(ctx, bson.M{"_id": self.user.Data().Id}, bson.M{"$inc": bson.M{"gaming.shit": how_many}})

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
