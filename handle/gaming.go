package handle

import (
	"encoding/json"
	"github.com/cosn/firebase"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/olebedev/config"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
)

type GamingAPI struct {
	DataService   *mongo.Service   `inject:""`
	CacheService  *goredis.Redis   `inject:""`
	ConfigService *config.Config   `inject:""`
	Firebase      *firebase.Client `inject:""`
	Errors        ErrorAPI         `inject:"inline"`
	Rules         *model.GamingRules
}

type GamingUserAPI struct {
	Services *GamingAPI
	UserId   bson.ObjectId
}

func (di *GamingAPI) Init() {

	// Get the redis interface from the DI
	redis := di.CacheService
	rules, _ := redis.Get("frontend.gamification")

	if rules == nil {

		gaming_file, err := di.ConfigService.String("application.gaming")

		if err != nil {
			panic(err)
		}

		rules_data, err := ioutil.ReadFile(gaming_file)

		// Unmarshal file with gaming rules
		if err := json.Unmarshal(rules_data, &di.Rules); err != nil {
			panic(err)
		}

		// Keep it warmed up
		err = redis.Set("frontend.gamification", string(rules_data), 43200, 0, false, false)

		if err != nil {
			panic(err)
		}

	} else {

		// Unmarshal already warmed up gaming rules
		if err := json.Unmarshal(rules, &di.Rules); err != nil {
			panic(err)
		}
	}
}

func (di *GamingAPI) ResetTempStuff() {

	// Recover from any panic even inside this goroutine
	defer di.Errors.Recover()

	var user *model.User

	// Get the database interface from the DI
	database := di.DataService.Database

	iter := database.C("users").Find(nil).Select(bson.M{"_id": 1}).Batch(500).Prefetch(0.25).Iter()

	log.Println("[job] [ResetTempStuff] Started")

	// Make a pool of workers that would execute the explote
	jobs := make(chan *model.User)

	worker := func(id int, jobs <-chan *model.User) {

		for j := range jobs {

			log.Printf("[job] [ResetTempStuff] [worker %v] User: %s\n", id, user.Id.Hex())

			// Explore the user level and reset the stuff
			di.Related(j.Id).ExploreRules(true)
		}
	}

	// Initialize the workers (25 concurrent)
	for w := 1; w <= 25; w++ {

		go worker(w, jobs)
	}

	for iter.Next(&user) {
		jobs <- user
	}

	close(jobs)

	if err := iter.Close(); err != nil {
		panic(err)
	}

	log.Println("[job] [ResetTempStuff] Finished")
}

func (di *GamingAPI) Related(user_id bson.ObjectId) *GamingUserAPI {

	user := &GamingUserAPI{
		Services: di,
		UserId:   user_id,
	}

	return user
}

func (di *GamingUserAPI) syncFirebase(data model.UserGaming) {

	// Get the user path from firebase
	userPath := "users/" + di.UserId.Hex()

	// Update the gaming part
	di.Services.Firebase.Set(userPath+"/gaming", data, nil)
}

func (di *GamingUserAPI) ExploreRules(reset bool) {

	var user model.User

	// Get the database interface from the DI
	database := di.Services.DataService.Database
	rules := di.Services.Rules.Rules

	// Get the user to sync
	err := database.C("users").Find(bson.M{"_id": di.UserId}).One(&user)

	if err != nil {
		return
	}

	// How many swords the user has
	user_swords := user.Gaming.Swords
	user_level := user.Gaming.Level

	for _, rule := range rules {

		if rule.Start <= user_swords && rule.End >= user_swords {

			// The user level just changed
			if user_level != rule.Level || reset {

				// Update the user
				changes := bson.M{"$set": bson.M{"gaming.level": rule.Level, "gaming.shit": rule.Shit, "gaming.tribute": rule.Tribute}}
				err := database.C("users").Update(bson.M{"_id": di.UserId}, changes)

				if err != nil {
					panic(err)
				}

				// Allows the inmediate syncing
				user.Gaming.Level = rule.Level
				user.Gaming.Shit = rule.Shit
				user.Gaming.Tribute = rule.Tribute

				di.syncFirebase(user.Gaming)
			}

			break
		}
	}
}

func (di *GamingUserAPI) Did(event string) *GamingUserAPI {

	// Recover from any panic even inside this goroutine
	defer di.Services.Errors.Recover()

	var swords int

	// Get the database interface from the DI
	database := di.Services.DataService.Database
	swords = 0

	// Usually the events are translated to swords in the gaming API
	if event == "comment" {

		swords = swords + 1
	} else if event == "publish" {

		swords = swords + 1
	}

	err := database.C("users").Update(bson.M{"_id": di.UserId}, bson.M{"$inc": bson.M{"gaming.swords": swords}})

	if err != nil {
		panic(err)
	}

	// Check for level changes and stuff
	di.ExploreRules(false)

	return di
}
