package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Id bson.ObjectId `bson:"_id,omitempty"`
}

func CalculateUsersStats(db *mgo.Database) {

	fmt.Println("Calculating user stats")

	users_collection := db.C("users")
	followers_collection := db.C("followers")
	posts_collection := db.C("posts")
	ranking_collection := db.C("ranking")

	var users []User

	// Attempt to fetch all the users
	err := users_collection.Find(bson.M{}).All(&users)

	if err != nil {

		panic(err)
	}

	for _, user := range users {

		var followers int
		var posts int

		// Attempt to get the followers of the gotten user
		followers, err = followers_collection.Find(bson.M{"following": user.Id}).Count()

		if err != nil {

			panic(err)
		}

		// Attempt to get the posts
		posts, err = posts_collection.Find(bson.M{"user_id": user.Id}).Count()

		if err != nil {

			panic(err)
		}

	}
}
