package main

import (
    "gopkg.in/mgo.v2/bson"
    "github.com/gin-gonic/gin"
)

type Category struct {
    Id bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
    Name string `bson:"name" json:"name"`
    Description string `bson:"description" json:"description"`
    Slug string `bson:"slug" json:"slug"`
    Permissions interface{} `bson:"permissions" json:"permissions"`
}

func CategoriesGet (c *gin.Context) {

    var categories []Category

    // Get the categories collection to perform there
    collection := database.C("categories")

    err := collection.Find(bson.M{}).All(&categories)

    if err != nil {
        panic(err)
    }

    c.JSON(200, categories)
}