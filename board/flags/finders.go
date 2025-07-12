package flags

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var FlagNotFound = errors.New("Flag has not been found by given criteria.")

func FindId(d deps, id primitive.ObjectID) (f Flag, err error) {
	ctx := context.TODO()
	err = d.Mgo().Collection("flags").FindOne(ctx, bson.M{"_id": id}).Decode(&f)
	if err == mongo.ErrNoDocuments {
		err = FlagNotFound
	}
	return
}

func FindOne(d deps, related string, relatedID, userID primitive.ObjectID) (f Flag, err error) {
	ctx := context.TODO()
	err = d.Mgo().Collection("flags").FindOne(ctx, bson.M{
		"related_to": related,
		"related_id": relatedID,
		"user_id":    userID,
	}).Decode(&f)
	if err == mongo.ErrNoDocuments {
		return f, FlagNotFound
	}

	return
}

func Count(d deps, q bson.M) int {
	ctx := context.TODO()
	n, err := d.Mgo().Collection("flags").CountDocuments(ctx, q)
	if err != nil {
		panic(err)
	}
	return int(n)
}

// TodaysCountByUser flags.
func TodaysCountByUser(d deps, id primitive.ObjectID) int {
	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, today.Location())
	return Count(d, bson.M{
		"user_id":    id,
		"created_at": bson.M{"$gte": startOfDay, "$lte": endOfDay},
	})
}
