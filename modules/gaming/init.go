package gaming

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/goinggo/work"
	"github.com/olebedev/config"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/exceptions"
	"github.com/tryanzu/core/modules/feed"
	"github.com/tryanzu/core/modules/user"
	"gopkg.in/mgo.v2/bson"
)

func Boot(file string) *Module {
	module := &Module{}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	// Unmarshal file with gaming rules
	if err := json.Unmarshal(data, &module.Rules); err != nil {
		panic(err)
	}

	return module
}

func logFunc(message string) {
	log.Println(message)
}

type Module struct {
	User   *user.Module                 `inject:""`
	Feed   *feed.FeedModule             `inject:""`
	Config *config.Config               `inject:""`
	Errors *exceptions.ExceptionsModule `inject:""`
	Rules  Rules
}

// Get user gaming struct
func (self *Module) Get(usr interface{}) *User {

	module := self

	switch usr.(type) {
	case bson.ObjectId:

		// Use user module reference to get the user and then create the user gaming instance
		user_obj, err := self.User.Get(usr.(bson.ObjectId))

		if err != nil {
			panic(err)
		}

		user_gaming := &User{user: user_obj, di: module}

		return user_gaming

	case *user.One:

		user_gaming := &User{user: usr.(*user.One), di: module}

		return user_gaming

	default:
		panic("Unkown argument")
	}
}

// Get post gaming struct
func (self *Module) Post(post interface{}) *Post {

	module := self

	switch post.(type) {
	case bson.ObjectId:

		// Use user module reference to get the user and then create the user gaming instance
		post_object, err := self.Feed.Post(post.(bson.ObjectId))

		if err != nil {
			panic(err)
		}

		post_gaming := &Post{post: post_object, di: module}

		return post_gaming

	case *feed.Post:

		post_gaming := &Post{post: post.(*feed.Post), di: module}

		return post_gaming

	default:
		panic("Unkown argument")
	}
}

// Get gamification model with badges
func (self *Module) GetRules() Rules {

	database := deps.Container.Mgo()
	rules := self.Rules

	err := database.C("badges").Find(nil).All(&rules.Badges)

	if err != nil {
		panic(err)
	}

	return rules
}

// Reset daily user stats
func (self *Module) ResetTempStuff() {

	// Recover from any panic even inside this goroutine
	defer self.Errors.Recover()

	var list []user.UserId

	w, err := work.New(5, time.Second, logFunc)

	if err != nil {
		log.Fatalln(err)
	}

	database := deps.Container.Mgo()
	count, _ := database.C("users").Find(nil).Count()
	err = database.C("users").Find(nil).Select(bson.M{"_id": 1}).All(&list)

	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	wg.Add(count)

	// Get the database interface from the DI
	log.Println("[job] [ResetTempStuff] Started")

	for _, usr := range list {

		usr_copy := usr
		user_sync := UserSync{
			user: usr_copy,
			gmf:  self.Get(usr_copy.Id),
		}

		w.Run(&user_sync)
		wg.Done()
	}

	wg.Wait()
	w.Shutdown()

	log.Printf("\n[job] [ResetTempStuff] Finished with users")
}

type UserSync struct {
	user user.UserId
	gmf  *User
}

func (self *UserSync) Work(id int) {

	log.Printf("\n user %v\n", self.user.Id)

	self.gmf.SyncToLevel(true)
}
