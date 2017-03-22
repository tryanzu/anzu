package deps

import (
	"gopkg.in/mgo.v2"
)

func IgniteMongoDB(container Deps) (Deps, error) {
	uri, err := container.Config.String("database.uri")
	if err != nil {
		return container, err
	}

	dbName, err := container.Config.String("database.name")
	if err != nil {
		return container, err
	}

	session, err := mgo.Dial(uri)
	if err != nil {
		return container, err
	}

	database := session.DB(dbName)

	// See https://godoc.org/gopkg.in/mgo.v2#Session.SetMode
	session.SetMode(mgo.Monotonic, true)

	container.Session = session
	container.Database = database

	return container, nil
}
