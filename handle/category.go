package handle

import (
	"encoding/json"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/gin-gonic/gin"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"strings"
)

type CategoryAPI struct {
	Acl          *acl.Module    `inject:""`
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

	err := collection.Find(nil).All(&categories)

	if err != nil {
		panic(err)
	}

	// Filter the parent categories without allocating
	parent_categories := []model.Category{}
	child_categories := []model.Category{}

	for _, category := range categories {

		if category.Parent.Hex() == "" {

			parent_categories = append(parent_categories, category)
		} else {

			child_categories = append(child_categories, category)
		}
	}

	for category_index, category := range parent_categories {

		for _, child := range child_categories {

			if child.Parent == category.Id {

				parent_categories[category_index].Child = append(parent_categories[category_index].Child, child)
			}
		}
	}

	// Check whether auth or not
	_, signed_in := c.Get("token")

	// Check the categories the user can write if parameter is provided
	perms := c.Query("permissions")

	if signed_in {

		user_id := c.MustGet("user_id").(string)

		if perms == "write" {

			user := di.Acl.User(bson.ObjectIdHex(user_id))

			// Remove child categories with no write permissions
			for category_index, _ := range parent_categories {

				for child_index, child := range parent_categories[category_index].Child {

					if user.CanWrite(child) == false {

						if len(parent_categories[category_index].Child) > child_index+1 {

							parent_categories[category_index].Child = append(parent_categories[category_index].Child[:child_index], parent_categories[category_index].Child[child_index+1:]...)
						} else {

							parent_categories[category_index].Child = parent_categories[category_index].Child[:len(parent_categories[category_index].Child)-1]
						}
					}
				}
			}

			// Clean up parent categories with no subcategories
			for category_index, category := range parent_categories {

				//log.Printf("%v has %v \n", category.Slug, len(category.Child))

				if len(category.Child) == 0 {

					if len(parent_categories) > category_index+1 {

						parent_categories = append(parent_categories[:category_index], parent_categories[category_index+1:]...)
					} else {

						parent_categories = parent_categories[:len(parent_categories)-1]
					}
				}
			}
		}

		user_counter := model.Counter{}
		err := database.C("counters").Find(bson.M{"user_id": user_id}).One(&user_counter)

		if err == nil {

			for category_index, category := range categories {

				counter_name := strings.Replace(category.Slug, "-", "_", -1)
				if _, okay := user_counter.Counters[counter_name]; okay {

					categories[category_index].Recent = user_counter.Counters[counter_name].Counter
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
					Slug:  category.Slug,
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

	sort.Sort(model.CategoriesOrder(parent_categories))

	c.JSON(200, parent_categories)
}
