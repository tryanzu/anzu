package assets

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FromURL asset.
func FromURL(deps Deps, url string) (ref Asset, err error) {
	ref = Asset{
		ID:       primitive.NewObjectID(),
		Original: url,
		Status:   "awaiting",
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = deps.Mgo().Collection("remote_assets").InsertOne(ctx, &ref)
	return
}
