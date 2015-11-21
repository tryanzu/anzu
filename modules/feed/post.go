package feed

import (
	"encoding/json"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/model"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strconv"
	"time"
)

type Post struct {
	di   *FeedModule
	data model.Post
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
		post: self,
		comment: self.data.Comments.Set[index],
		index: index,
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
		add, _ := helpers.InArray(id, self.data.RelatedComponents)

		if add {

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
