package deps

import (
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

	session, err := mgo.Dial(uri)
	if err != nil {
		return container, err
	}

	database := session.DB(dbName)

	// Ensure indexes
	search := mgo.Index{
		Key: []string{"$text:title", "$text:content"},
		Weights: map[string]int{
			"title":   3,
			"content": 1,
		},
		DefaultLanguage: "spanish",
		Background:      true, // See notes.
	}
	database.C("posts").EnsureIndex(search)

	// See https://godoc.org/gopkg.in/mgo.v2#Session.SetMode
	//session.SetMode(mgo.Monotonic, true)

	container.DatabaseSessionProvider = session
	container.DatabaseProvider = database

	return container, nil
}
