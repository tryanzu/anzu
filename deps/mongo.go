package deps

import (
	"gopkg.in/mgo.v2"
)

var (
	// MongoURL config uri
	MongoURL string
	// MongoName db name
	MongoName string
)

func IgniteMongoDB(container Deps) (Deps, error) {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Error(err)
		log.Info(MongoURL)
		return container, err
	}
	db := session.DB(MongoName)

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
