package main

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"sort"
	"encoding/json"
)

type Category struct {
	Id          bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	Slug        string        `bson:"slug" json:"slug"`
	Permissions interface{}   `bson:"permissions" json:"permissions"`
	Count       int 		  `bson:"count,omitempty" json:"count,omitempty"`
	Recent      int           `bson:"recent,omitempty" json:"recent,omitempty"`
}

type CategoryCounters struct {
	List 		[]CategoryCounter `json:"list"`
}

type CategoryCounter struct {
	Slug 		string `json:"slug"`
	Count 		int    `json:"count"`
}


type Categories []Category

func (slice Categories) Len() int {
    return len(slice)
}

func (slice Categories) Less(i, j int) bool {
    return slice[i].Count > slice[j].Count;
}

func (slice Categories) Swap(i, j int) {
    slice[i], slice[j] = slice[j], slice[i]
}

func CategoriesGet(c *gin.Context) {

	var categories []Category

	// Get the categories collection to perform there
	collection := database.C("categories")

	err := collection.Find(bson.M{}).All(&categories)

	if err != nil {
		panic(err)
	}

	// Check whether auth or not
	user_token := UserToken{}
	token := c.Request.Header.Get("Auth-Token")

	if token != "" {

		// Try to fetch the user using token header though
		err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)

		if err == nil {

			user_counter := Counter{}
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

		list := []CategoryCounter{}

		for category_index, category := range categories {

			count, err := database.C("posts").Find(bson.M{"categories": category.Slug}).Count()

			if err == nil {

				// Append to the list of counters
				list = append(list, CategoryCounter{
					Slug: category.Slug,
					Count: count,
				})

				// Add the count to the current set of categories
				categories[category_index].Count = count
			}
		}

		cache := CategoryCounters{
			List: list,
		}

		encoded, err := json.Marshal(cache)

		err = redis.Set("frontend.counters.category", string(encoded), 43200, 0, false, false)

		if err != nil {
			panic(err)
		}

	} else {

		var cache CategoryCounters

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

	sort.Sort(Categories(categories))

	c.JSON(200, categories)
}
