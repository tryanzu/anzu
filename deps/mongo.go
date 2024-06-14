package deps

import (
	"flag"

	"gopkg.in/mgo.v2"
)

var (
	// MongoURL config uri
	MongoURL string
	// MongoName db name
	MongoName string

	ShouldSeed = flag.Bool("should-seed", false, "determines whether we seed the initial categories and admin user to bootstrap the site")
)

func IgniteMongoDB(container Deps) (Deps, error) {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Error(err)
		log.Info(MongoURL)
		return container, err
	}
	db := session.DB(MongoName)
	collections, err := db.CollectionNames()
	if err != nil {
		return container, err
	}
	seed := true
	for _, v := range collections {
		if v == "users" {
			seed = false
			break
		}
	}
	if seed {
		ShouldSeed = &seed
	}
	// Ensure indexes
	db.C("users").EnsureIndex(
		mgo.Index{
			Key:        []string{"email"},
			Unique:     true,
			Background: true,
		},
	)
	db.C("users").EnsureIndex(
		mgo.Index{
			Key:        []string{"username"},
			Unique:     true,
			Background: true,
		},
	)
	search := mgo.Index{
		Key: []string{"$text:title", "$text:content"},
		Weights: map[string]int{
			"title":   3,
			"content": 1,
		},
		DefaultLanguage: "spanish",
		Background:      true, // See notes.
	}
	db.C("posts").EnsureIndex(search)

	// See https://godoc.org/gopkg.in/mgo.v2#Session.SetMode
	//session.SetMode(mgo.Monotonic, true)

	container.DatabaseSessionProvider = session
	container.DatabaseProvider = db

	return container, nil
}
