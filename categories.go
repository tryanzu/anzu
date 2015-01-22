package main

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

type Category struct {
	Id          bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	Slug        string        `bson:"slug" json:"slug"`
	Permissions interface{}   `bson:"permissions" json:"permissions"`
	Recent      int           `bson:"recent,omitempty" json:"recent,omitempty"`
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

	c.JSON(200, categories)
}
