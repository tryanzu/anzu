package deps

import (
	"os"

	"gopkg.in/mgo.v2"
)

func IgniteMongoDB(container Deps) (Deps, error) {
	uri, err := container.Config().String("database.uri")
	if err != nil {
		return container, err
	}

	dbName, err := container.Config().String("database.name")
	if err != nil {
		return container, err
	}
	if envURI := os.Getenv("MONGO_URL"); len(envURI) > 0 {
		uri = envURI
	}
	session, err := mgo.Dial(uri)
	if err != nil {
		return container, err
	}
	if envDb := os.Getenv("MONGO_DB"); len(envDb) > 0 {
		dbName = envDb
	}
	db := session.DB(dbName)

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
