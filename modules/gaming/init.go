package gaming

import (
	"encoding/json"
	"github.com/cosn/firebase"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/olebedev/config"
	"github.com/goinggo/work"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"time"
	"sync"
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

func logFunc(message string) {
    log.Println(message)
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

	w, err := work.New(40, time.Second, logFunc)

	if err != nil {
        log.Fatalln(err)
    }

	database := self.Mongo.Database
	count, _ := database.C("users").Find(nil).Count()
	iter := database.C("users").Find(nil).Select(bson.M{"_id": 1}).Iter()
    
    var wg sync.WaitGroup

    wg.Add(count)

	var usr user.User

	// Get the database interface from the DI
	log.Println("[job] [ResetTempStuff] Started")

	for iter.Next(&usr) {

		user_sync := UserSync{
			user: usr,
			gmf: self.Get(usr.Id),
		}
		
		go func() {

			w.Run(&user_sync)
			wg.Done()
		}()
	}

	if err := iter.Close(); err != nil {
		panic(err)
	}

	wg.Wait()
	w.Shutdown()

	log.Printf("\n[job] [ResetTempStuff] Finished with users")
}

type UserSync struct {
	user user.User
	gmf  *User
}

func (self *UserSync) Work(id int) {

	log.Printf("\n user %v\n", self.user.Id)

	self.gmf.SyncToLevel(true)
}