package feed

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"gopkg.in/mgo.v2/bson"
	"time"
	"strconv"
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
		UserId: user_id, 
		Event: "post", 
		RelatedId: self.data.Id,
		Created: time.Now(),
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

func (self *Post) Data() model.Post {
	return self.data
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
