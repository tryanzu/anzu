package feed

import (
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/board/votes"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/exceptions"
	"github.com/tryanzu/core/modules/helpers"
	"github.com/tryanzu/core/modules/user"
	"gopkg.in/mgo.v2/bson"

	"fmt"
	"html"
	"strconv"
	"time"
)

// Post model refers to board posts
type Post struct {
	Id                bson.ObjectId    `bson:"_id,omitempty" json:"id,omitempty"`
	Title             string           `bson:"title" json:"title"`
	Slug              string           `bson:"slug" json:"slug"`
	Type              string           `bson:"type" json:"type"`
	Content           string           `bson:"content" json:"content"`
	Categories        []string         `bson:"categories" json:"categories"`
	Category          bson.ObjectId    `bson:"category" json:"category"`
	Comments          Comments         `bson:"comments" json:"comments"`
	Author            *user.UserSimple `bson:"-" json:"author,omitempty"`
	UserId            bson.ObjectId    `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Users             []bson.ObjectId  `bson:"users,omitempty" json:"users,omitempty"`
	Votes             votes.Votes      `bson:"votes" json:"votes"`
	RelatedComponents []bson.ObjectId  `bson:"related_components,omitempty" json:"related_components,omitempty"`
	Following         bool             `bson:"following,omitempty" json:"following,omitempty"`
	Pinned            bool             `bson:"pinned,omitempty" json:"pinned,omitempty"`
	Lock              bool             `bson:"lock" json:"lock"`
	IsQuestion        bool             `bson:"is_question" json:"is_question"`
	Solved            bool             `bson:"solved,omitempty" json:"solved,omitempty"`
	Voted             []string         `bson:"-" json:"voted"`
	Created           time.Time        `bson:"created_at" json:"created_at"`
	Updated           time.Time        `bson:"updated_at" json:"updated_at"`
	Deleted           time.Time        `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`

	// Runtime generated pointers
	UsersHashtable map[string]interface{} `bson:"-" json:"usersHashtable"`
	di             *FeedModule
}

// Set Dependency Injection pointer
func (self *Post) SetDI(di *FeedModule) {
	self.di = di
}

func (self *Post) LoadUsersHashtables() {
	var users []user.UserSimple
	ids := self.Users
	ids = append(ids, self.UserId)
	err := deps.Container.Mgo().C("users").Find(bson.M{"_id": bson.M{"$in": ids}}).All(&users)
	if err != nil {
		panic(err)
	}

	ht := make(map[string]interface{}, len(users))
	for _, u := range users {
		ht[u.UserName] = u
	}

	self.UsersHashtable = ht
	return
}

// Comment loading by ID for post
func (self *Post) LoadCommentById(id bson.ObjectId) error {

	var c *Comment

	// Use content module to run processors chain
	content := self.di.Content
	database := deps.Container.Mgo()
	err := database.C("comments").Find(bson.M{"_id": id, "deleted_at": bson.M{"$exists": false}}).One(&c)

	if err != nil {
		self.Comments.Set = make([]*Comment, 0)
		return err
	}

	self.Comments.Set = []*Comment{c}

	for _, comment := range self.Comments.Set {
		content.ParseTags(comment)
	}

	return nil
}

// Push Comment on the post
func (self *Post) PushComment(c string, user_id bson.ObjectId) *Comment {

	c = html.EscapeString(c)
	if len(c) > 3000 {
		c = c[:3000] + "..."
	}

	pos := self.GetCommentCount()

	comment := &Comment{
		Id:       bson.NewObjectId(),
		PostId:   self.Id,
		UserId:   user_id,
		Content:  c,
		Position: pos,
		Created:  time.Now(),
	}

	comment.SetDI(self)

	// Use content module to run processors chain
	content := self.di.Content
	content.Parse(comment)

	// Publish comment
	database := deps.Container.Mgo()
	err := database.C("comments").Insert(comment)
	if err != nil {
		panic(err)
	}

	err = database.C("posts").Update(bson.M{"_id": self.Id}, bson.M{"$set": bson.M{"updated_at": time.Now()}, "$inc": bson.M{"comments.count": 1}})
	if err != nil {
		panic(err)
	}

	// Ensure user is participating
	self.PushUser(user_id)

	// Finally parse tags in content for runtime usage
	content.ParseTags(comment)

	return comment
}

// Push new user to the participants list
func (p *Post) PushUser(user_id bson.ObjectId) bool {

	pushed := false

	for _, u := range p.Users {

		if u == user_id {
			pushed = true
		}
	}

	if !pushed {
		err := deps.Container.Mgo().C("posts").Update(bson.M{"_id": p.Id}, bson.M{"$push": bson.M{"users": user_id}})

		if err != nil {
			panic(err)
		}
	}

	return pushed
}

// Eager load users for post entities
func (self *Post) LoadUsers() {

	var list []bson.ObjectId
	var users []user.UserSimple

	// Check if author need to be loaded
	if self.Author == nil {
		list = append(list, self.UserId)
	}

	// Load comment set authors at runtime
	if len(self.Comments.Set) > 0 {
		for _, c := range self.Comments.Set {

			// Do not repeat ids at the list
			if exists, _ := helpers.InArray(c.UserId, list); !exists {
				list = append(list, c.UserId)
			}
		}

		// Best answer author
		if self.Comments.Answer != nil {
			if exists, _ := helpers.InArray(self.Comments.Answer.UserId, list); !exists {
				list = append(list, self.Comments.Answer.UserId)
			}
		}
	}

	if len(list) > 0 {

		database := deps.Container.Mgo()
		err := database.C("users").Find(bson.M{"_id": bson.M{"$in": list}}).Select(user.UserSimpleFields).All(&users)

		if err != nil {
			panic(err)
		}

		usersMap := make(map[bson.ObjectId]interface{})

		for i, usr := range users {

			if len(usr.Description) == 0 {
				usr.Description = "Solo otro Spartan Geek más"
			}

			usersMap[usr.Id] = usr

			if usr.Id == self.UserId {

				fmt.Println("User author is", usr.Id.Hex(), self.UserId.Hex())

				self.Author = &users[i]
			}

			if self.Comments.Answer != nil && self.Comments.Answer.UserId == usr.Id {
				self.Comments.Answer.User = usersMap[usr.Id]
			}
		}

		for index, c := range self.Comments.Set {
			if usr, exists := usersMap[c.UserId]; exists {
				self.Comments.Set[index].User = usr
			}
		}
	}
}

// Load voting status for certain user
func (self *Post) LoadVotes(user_id bson.ObjectId) {
	var list []votes.Vote
	err := deps.Container.Mgo().C("votes").Find(bson.M{
		"type":       "post",
		"related_id": self.Id,
		"user_id":    user_id,
		"deleted_at": bson.M{"$exists": false},
	}).All(&list)
	if err != nil {
		panic(err)
	}
	self.Voted = make([]string, 0)
	for _, v := range list {
		self.Voted = append(self.Voted, v.Value)
	}
}

// Collects the post views
func (self *Post) Viewed(user_id bson.ObjectId) {

	database := deps.Container.Mgo()
	redis := self.di.CacheService

	activity := model.Activity{
		UserId:    user_id,
		Event:     "post",
		RelatedId: self.Id,
		Created:   time.Now(),
	}

	err := database.C("activity").Insert(activity)

	if err != nil {
		panic(err)
	}

	// Increase the post views inside the cache service
	viewed_count, _ := redis.Get("feed:count:post:" + self.Id.Hex())

	if viewed_count != nil {

		_, err := redis.Incr("feed:count:post:" + self.Id.Hex())

		if err != nil {
			panic(err)
		}
	} else {

		// No need to get the numbers but to warm up cache
		self.GetReachViews(self.Id)
	}
}

// Update ranking rates
func (self *Post) UpdateRate() {

	// Services we will need along the runtime
	redis := self.di.CacheService

	// Sorted list items (redis ZADD)
	zadd := make(map[string]float64)

	// Get reach and views
	reached, viewed := self.GetReachViews(self.Id)

	total := reached + viewed

	if total > 101 {

		// Calculate the rates
		view_rate := 100.0 / float64(reached) * float64(viewed)
		comment_rate := 100.0 / float64(viewed) * float64(self.Comments.Count)
		final_rate := (view_rate + comment_rate) / 2.0
		date := self.Created.Format("2006-01-02")

		zadd[self.Id.Hex()] = final_rate

		_, err := redis.ZAdd("feed:relevant:"+date, zadd)

		if err != nil {
			panic(err)
		}
	}
}

// Get post data structure
func (self *Post) Data() *Post {
	return self
}

func (self *Post) IsLocked() bool {
	return self.Lock
}

func (self *Post) DI() *FeedModule {
	return self.di
}

// Internal method to get the post reach and views - TODO: move this up in the hierarchy
func (self *Post) GetReachViews(id bson.ObjectId) (int, int) {

	var reached, viewed int

	// Services we will need along the runtime
	database := deps.Container.Mgo()
	redis := self.di.CacheService

	list_count, _ := redis.Get("feed:count:list:" + id.Hex())

	if list_count == nil {

		reached, _ = database.C("activity").Find(bson.M{"list": id, "event": "feed"}).Count()
		err := redis.Set("feed:count:list:"+id.Hex(), strconv.Itoa(reached), 1800, 0, false, false)

		if err != nil {
			panic(err)
		}
	} else {

		reached, _ = strconv.Atoi(string(list_count))
	}

	viewed_count, _ := redis.Get("feed:count:post:" + id.Hex())

	if viewed_count == nil {

		viewed, _ = database.C("activity").Find(bson.M{"related_id": id, "event": "post"}).Count()
		err := redis.Set("feed:count:post:"+id.Hex(), strconv.Itoa(viewed), 1800, 0, false, false)

		if err != nil {
			panic(err)
		}
	} else {

		viewed, _ = strconv.Atoi(string(viewed_count))
	}

	return reached, viewed
}

// Get post category model - TODO: determine usage cases
func (self *Post) GetCategory() model.Category {

	var category model.Category

	database := deps.Container.Mgo()
	err := database.C("categories").Find(bson.M{"_id": self.Category}).One(&category)

	if err != nil {
		panic(err)
	}

	return category
}

// Get comment object
func (self *Post) Comment(index int) (*Comment, error) {

	var comment *Comment

	database := deps.Container.Mgo()
	err := database.C("comments").Find(bson.M{"post_id": self.Id, "position": index, "deleted_at": bson.M{"$exists": false}}).One(&comment)

	if err != nil {
		return nil, exceptions.OutOfBounds{"Invalid comment index"}
	}

	comment.SetDI(self)

	return comment, nil
}

// Alias of TrueCommentCount
func (self *Post) GetCommentCount() int {
	return self.di.TrueCommentCount(self.Id)
}
