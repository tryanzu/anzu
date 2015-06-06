package handle

import (
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/gin-gonic/gin"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"sort"
	"encoding/json"
)

type CategoryAPI struct {
	DataService  *mongo.Service `inject:""`
	CacheService *goredis.Redis `inject:""`
}

func (di *CategoryAPI) CategoriesGet(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database
	redis := di.CacheService

	var categories []model.Category

	// Get the categories collection to perform there
	collection := database.C("categories")

	err := collection.Find(bson.M{}).All(&categories)

	if err != nil {
		panic(err)
	}

	// Check whether auth or not
	user_token := model.UserToken{}
	token := c.Request.Header.Get("Auth-Token")

	if token != "" {

		// Try to fetch the user using token header though
		err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)

		if err == nil {

			user_counter := model.Counter{}
			err := database.C("counters").Find(bson.M{"user_id": user_token.UserId}).One(&user_counter)

			if err == nil {

				for category_index, category := range categories {

					counter_name := strings.Replace(category.Slug, "-", "_", -1)
					if _, okay := user_counter.Counters[counter_name]; okay {

						categories[category_index].Recent = user_counter.Counters[counter_name].Counter
					}
				}
			}
		}
	}

	counters, _ := redis.Get("frontend.counters.category")

	if counters == nil {

		list := []model.CategoryCounter{}

		for category_index, category := range categories {

			count, err := database.C("posts").Find(bson.M{"categories": category.Slug}).Count()

			if err == nil {

				// Append to the list of counters
				list = append(list, model.CategoryCounter{
					Slug: category.Slug,
					Count: count,
				})

				// Add the count to the current set of categories
				categories[category_index].Count = count
			}
		}

		cache := model.CategoryCounters{
			List: list,
		}

		encoded, err := json.Marshal(cache)

		err = redis.Set("frontend.counters.category", string(encoded), 43200, 0, false, false)

		if err != nil {
			panic(err)
		}

	} else {

		var cache model.CategoryCounters

		// Unmarshal already warmed up user achievements
        if err := json.Unmarshal(counters, &cache); err != nil {
            panic(err)
        }

        for _, category := range cache.List {

        	for current_index, current := range categories {

        		if current.Slug == category.Slug {

        			// Set the current count of the category
        			categories[current_index].Count = category.Count
        		}
        	}
        }
	}

	sort.Sort(model.Categories(categories))

	c.JSON(200, categories)
}
