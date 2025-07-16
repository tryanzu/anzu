package flags

import (
	"context"
	"html"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UpsertComment performs validations before upserting data struct
func UpsertFlag(d DepsInterface, f Flag) (flag Flag, err error) {
	ctx := context.TODO()
	if f.ID.IsZero() {
		f.ID = primitive.NewObjectID()
		f.Created = time.Now()
		f.Status = PENDING
	}

	f.Content = html.EscapeString(f.Content)
	f.Updated = time.Now()
	upsertTrue := true
	_, err = d.Mgo().Collection("flags").ReplaceOne(ctx, bson.M{"_id": f.ID}, f, &options.ReplaceOptions{Upsert: &upsertTrue})
	if err != nil {
		return
	}

	flag = f
	return
}
