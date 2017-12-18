package handle

import (
	"encoding/json"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/gin-gonic/gin"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
)

type StatAPI struct {
	CacheService *goredis.Redis `inject:""`
}

func (di *StatAPI) BoardGet(c *gin.Context) {

	// Get the database interface from the DI
	database := deps.Container.Mgo()
	redis := di.CacheService

	stats, _ := redis.Get("frontend.stats")

	if stats == nil {

		var users int
		var posts int
		var comments model.StatsComments

		// Get the posts collection to perform there
		collection := database.C("posts")

		pipe := collection.Pipe([]bson.M{{"$group": bson.M{"_id": "", "count": bson.M{"$sum": "$comments.count"}}}})
		err := pipe.One(&comments)

		if err != nil {
			panic(err)
		}

		users, err = database.C("users").Find(bson.M{}).Count()
		posts, err = database.C("posts").Find(bson.M{}).Count()

		stat := model.Stats{
			Comments: comments.Count,
			Users:    users,
			Posts:    posts,
		}

		encoded, err := json.Marshal(stat)
		err = redis.Set("frontend.stats", string(encoded), 600, 0, false, false)

		if err != nil {
			panic(err)
		}

		c.JSON(200, stat)

	} else {

		var cache model.Stats

		// Unmarshal already warmed up user achievements
		if err := json.Unmarshal(stats, &cache); err != nil {
			panic(err)
		}

		c.JSON(200, cache)
	}
}
