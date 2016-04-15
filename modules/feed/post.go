package feed

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"gopkg.in/mgo.v2/bson"

	"encoding/json"
	"reflect"
	"strconv"
	"time"
)

type Post struct {
	di   *FeedModule
	data model.Post
}

// Comments loading for post
func (self *Post) LoadComments(take, skip int) {

	var c []model.Comment
	var limit int = take
	var sort string

	database := self.di.Mongo.Database

	if take > 0 {
		sort = "created_at"
	} else {
		sort = "-created_at"
		limit = -take
	}

	err := database.C("comments").Find(bson.M{"post_id": self.data.Id, "deleted_at": bson.M{"$exists": false}}).Sort(sort).Skip(skip).Limit(limit).All(&c)

	if err != nil {
		panic(err)
	}

	self.data.Comments.Set = c

	// Load the best answer if needed
	if self.data.Solved == true && self.data.Comments.Answer == nil {

		loaded := false

		// We may have the chosen answer within the loaded comments set
		for _, c := range self.data.Comments.Set {
			if c.Chosen == true {
				loaded = true
				self.data.Comments.Answer = &c
			}
		}

		if !loaded {

			var ca model.Comment

			// Load the chosen answer from Database
			err := database.C("comments").Find(bson.M{"post_id": self.data.Id, "deleted_at": bson.M{"$exists": false}, "chosen": true}).One(&ca)

			if err != nil {
				panic(err)
			}

			self.data.Comments.Answer = &ca
		}
	}
}

func (self *Post) LoadUsers() {

	var list []bson.ObjectId
	var users []model.User

	// Check if author need to be loaded
	if !self.data.Author.Id.Valid() {
		list = append(list, self.data.UserId)
	}

	// Load comment set authors at runtime
	if len(self.data.Comments.Set) > 0 {
		for _, c := range self.data.Comments.Set {

			// Do not repeat ids at the list
			if exists, _ := helpers.InArray(c.UserId, list); !exists {
				list = append(list, c.UserId)
			}
		}

		// Best answer author
		if self.data.Comments.Answer != nil {

			if exists, _ := helpers.InArray(self.data.Comments.Answer.UserId, list); !exists {
				list = append(list, self.data.Comments.Answer.UserId)
			}
		}
	}

	if len(list) > 0 {

		database := self.di.Mongo.Database
		err := database.C("users").Find(bson.M{"_id": bson.M{"$in": list}}).All(&users)

		if err != nil {
			panic(err)
		}

		usersMap := make(map[bson.ObjectId]interface{})

		for _, user := range users {

			description := "Solo otro Spartan Geek más"

			if len(user.Description) > 0 {
				description = user.Description
			}

			usersMap[user.Id] = map[string]interface{}{
				"id":          user.Id.Hex(),
				"username":    user.UserName,
				"description": description,
				"image":       user.Image,
				"level":       user.Gaming.Level,
				"roles":       user.Roles,
			}

			if user.Id == self.data.UserId {
				self.data.Author = user
			}

			if self.data.Comments.Answer != nil && self.data.Comments.Answer.UserId == user.Id {
				self.data.Comments.Answer.User = usersMap[user.Id]
			}
		}

		for index, c := range self.data.Comments.Set {

			if user, exists := usersMap[c.UserId]; exists {

				self.data.Comments.Set[index].User = user
			}
		}
	}
}

func (self *Post) LoadVotes(user_id bson.ObjectId) {

	var post model.Vote
	database := self.di.Mongo.Database

	// Only when there's loaded comments on the post
	if len(self.data.Comments.Set) > 0 {

		var positions []string
		var comments []model.Vote

		for _, c := range self.data.Comments.Set {

			// Transform to string due to the fact that nested type inside votes used to be strings - TODO: normalize
			p := strconv.Itoa(c.Position)
			positions = append(positions, p)
		}

		// Get votes given to post's comments
		err := database.C("votes").Find(bson.M{"type": "comment", "related_id": self.data.Id, "nested_type": bson.M{"$in": positions}, "user_id": user_id}).All(&comments)

		if err != nil {
			panic(err)
		}

		if len(comments) > 0 {

			for index, c := range self.data.Comments.Set {

				p := strconv.Itoa(c.Position)

				// Iterate over comment's votes to determine status
				for _, v := range comments {

					if v.NestedType == p {
						self.data.Comments.Set[index].Liked = v.Value
					}
				}
			}
		}
	}

	err := database.C("votes").Find(bson.M{"type": "post", "related_id": post.Id, "user_id": user_id}).One(&post)

	if err == nil {
		self.data.Liked = post.Value
	}
}

// Collects the post views
func (self *Post) Viewed(user_id bson.ObjectId) {

	database := self.di.Mongo.Database
	redis := self.di.CacheService

	activity := model.Activity{
		UserId:    user_id,
		Event:     "post",
		RelatedId: self.data.Id,
		Created:   time.Now(),
	}

	err := database.C("activity").Insert(activity)

	if err != nil {
		panic(err)
	}

	// Increase the post views inside the cache service
	viewed_count, _ := redis.Get("feed:count:post:" + self.data.Id.Hex())

	if viewed_count != nil {

		_, err := redis.Incr("feed:count:post:" + self.data.Id.Hex())

		if err != nil {
			panic(err)
		}
	} else {

		// No need to get the numbers but to warm up cache
		self.GetReachViews(self.data.Id)
	}
}

// Update ranking rates
func (self *Post) UpdateRate() {

	// Services we will need along the runtime
	redis := self.di.CacheService

	// Sorted list items (redis ZADD)
	zadd := make(map[string]float64)

	// Get reach and views
	reached, viewed := self.GetReachViews(self.data.Id)

	total := reached + viewed

	if total > 101 {

		// Calculate the rates
		view_rate := 100.0 / float64(reached) * float64(viewed)
		comment_rate := 100.0 / float64(viewed) * float64(self.data.Comments.Count)
		final_rate := (view_rate + comment_rate) / 2.0
		date := self.data.Created.Format("2006-01-02")

		zadd[self.data.Id.Hex()] = final_rate

		_, err := redis.ZAdd("feed:relevant:"+date, zadd)

		if err != nil {
			panic(err)
		}
	}
}

// Get post data structure
func (self *Post) Data() model.Post {
	return self.data
}

func (self *Post) IsLocked() {
	return self.data.Locked
}

func (self *Post) DI() *FeedModule {
	return self.di
}

// Internal method to get the post reach and views
func (self *Post) GetReachViews(id bson.ObjectId) (int, int) {

	var reached, viewed int

	// Services we will need along the runtime
	database := self.di.Mongo.Database
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

// Get post category model
func (self *Post) Category() model.Category {

	var category model.Category

	database := self.di.Mongo.Database
	err := database.C("categories").Find(bson.M{"_id": self.data.Category}).One(&category)

	if err != nil {
		panic(err)
	}

	return category
}

// Get comment object
func (self *Post) Comment(index int) (*Comment, error) {

	if len(self.data.Comments.Set) < index {

		return nil, exceptions.OutOfBounds{"Invalid comment index"}
	}

	comment := &Comment{
		post:    self,
		comment: self.data.Comments.Set[index],
		index:   index,
	}

	return comment, nil
}

// Attach related entity to post
func (self *Post) Attach(entity interface{}) {

	database := self.di.Mongo.Database

	switch entity.(type) {
	case *components.ComponentModel:

		component := entity.(*components.ComponentModel)
		id := component.Id

		// Check if we need to relate the component
		exists, _ := helpers.InArray(id, self.data.RelatedComponents)

		if !exists {

			err := database.C("posts").Update(bson.M{"_id": self.data.Id}, bson.M{"$push": bson.M{"related_components": id}})

			if err != nil {
				panic(err)
			}
		}

	default:
		panic("Unkown argument")
	}
}

// Use algolia to index the post
func (self *Post) Index() {

	if false {

		post := self.data

		if post.Category.Hex() != "" {

			index := self.di.Search.Get("board")
			category := self.Category()
			user, err := self.di.User.Get(post.UserId)

			if err != nil {
				panic(err)
			}

			// Some data we do use on searches
			user_data := user.Data()
			tribute := post.Votes.Up
			shit := post.Votes.Down

			for _, comment := range post.Comments.Set {

				tribute = tribute + comment.Votes.Up
				shit = shit + comment.Votes.Down
			}

			components := make([]string, 0)

			// If the post is a recommendations post then reflect to get the components
			if post.Type == "recommendations" {

				bindable := reflect.ValueOf(&post.Components).Elem()

				for i := 0; i < bindable.NumField(); i++ {

					field := bindable.Field(i).Interface()

					switch field.(type) {
					case model.Component:

						component := field.(model.Component)

						if component.Content != "" {

							components = append(components, component.Content)
						}

					default:
						continue
					}
				}
			}

			reached, viewed := self.GetReachViews(post.Id)
			total := reached + viewed
			final_rate := 0.0

			if total > 101 {

				if reached == 0 {

					reached = 1
				}

				if viewed == 0 {

					viewed = 1
				}

				view_rate := 100.0 / float64(reached) * float64(viewed)
				comment_rate := 100.0 / float64(viewed) * float64(post.Comments.Count)
				final_rate = (view_rate + comment_rate) / 2.0
			}

			item := AlgoliaPostModel{
				Id:       post.Id.Hex(),
				Title:    post.Title,
				Content:  post.Content,
				Slug:     post.Slug,
				Comments: post.Comments.Count,
				User: AlgoliaUserModel{
					Id:       user_data.Id.Hex(),
					Username: user_data.UserName,
					Image:    user_data.Image,
					Email:    user_data.Email,
				},
				Category: AlgoliaCategoryModel{
					Id:   post.Category.Hex(),
					Name: category.Name,
				},
				Popularity: final_rate,
				Created:    post.Created.Unix(),
				Components: components,
			}

			var json_object interface{}
			json_data, err := json.Marshal(item)

			if err != nil {
				panic(err)
			}

			err = json.Unmarshal(json_data, &json_object)

			if err != nil {
				panic(err)
			}

			_, err = index.UpdateObject(json_object)

			if err != nil {
				panic(err)
			}
		}
	}
}
