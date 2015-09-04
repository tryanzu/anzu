package gaming

import (
	"encoding/json"
	"github.com/cosn/firebase"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/olebedev/config"
	"github.com/jeffail/tunny"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
)

func Boot(file string) *Module {

	module := &Module{}
	rules_data, err := ioutil.ReadFile(file)

	if err != nil {
		panic(err)
	}

	// Unmarshal file with gaming rules
	if err := json.Unmarshal(rules_data, &module.Rules); err != nil {
		panic(err)
	}

	return module
}

type Module struct {
	Mongo    *mongo.Service               `inject:""`
	User     *user.Module                 `inject:""`
	Redis    *goredis.Redis               `inject:""`
	Config   *config.Config               `inject:""`
	Firebase *firebase.Client             `inject:""`
	Errors   *exceptions.ExceptionsModule `inject:""`
	Rules    model.GamingRules
}

// Get user gaming struct
func (self *Module) Get(usr interface{}) *User {

	module := self

	switch usr.(type) {
	case bson.ObjectId:

		// Use user module reference to get the user and then create the user gaming instance
		user_obj := self.User.Get(usr.(bson.ObjectId))
		user_gaming := &User{user: user_obj, di: module}

		return user_gaming

	case *user.One:

		user_gaming := &User{user: usr.(*user.One), di: module}

		return user_gaming

	default:
		panic("Unkown argument")
	}
}

func (self *Module) ResetTempStuff() {

	// Recover from any panic even inside this goroutine
	defer self.Errors.Recover()

	var usr user.User

	// Get the database interface from the DI
	database := self.Mongo.Database
	iter := database.C("users").Find(nil).Select(bson.M{"_id": 1}).Iter()

	log.Println("[job] [ResetTempStuff] Started")

	pool, _ := tunny.CreatePool(20, func(p interface{}) interface{} {
        
		j := p.(user.User)
        selfcopy := self

		log.Printf("[job] [ResetTempStuff] User: %s\n", j.Id.Hex())

		// Explore the user level and reset the stuff
		selfcopy.Get(j.Id).SyncToLevel(true)

		return true

    }).Open()

    defer pool.Close()

	for iter.Next(&usr) {
		
		pool.SendWorkAsync(usr, nil)
	}

	if err := iter.Close(); err != nil {
		panic(err)
	}

	log.Printf("\n[job] [ResetTempStuff] Finished with users")
}
