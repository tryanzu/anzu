package flags

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2/bson"
)

var FlagNotFound = errors.New("Flag has not been found by given criteria.")

func FindId(d deps, id bson.ObjectId) (f Flag, err error) {
	err = d.Mgo().C("flags").FindId(id).One(&f)
	if err != nil {
		err = FlagNotFound
	}
	return
}

func FindOne(d deps, related string, relatedID, userID bson.ObjectId) (f Flag, err error) {
	err = d.Mgo().C("flags").Find(bson.M{
		"related_to": related,
		"related_id": relatedID,
		"user_id":    userID,
	}).One(&f)
	if err != nil {
		return f, FlagNotFound
	}

	return
}

func Count(d deps, q bson.M) int {
	n, err := d.Mgo().C("flags").Find(q).Count()
	if err != nil {
		panic(err)
	}
	return n
}

// TodaysCountByUser flags.
func TodaysCountByUser(d deps, id bson.ObjectId) int {
	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, today.Location())
	return Count(d, bson.M{
		"user_id":    id,
		"created_at": bson.M{"$gte": startOfDay, "$lte": endOfDay},
	})
}
