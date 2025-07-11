package notifications

import (
	"context"
	"errors"

	"github.com/tryanzu/core/core/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var NotificationNotFound = errors.New("Notification has not been found by given criteria.")

func FindId(deps Deps, id primitive.ObjectID) (notification Notification, err error) {
	err = deps.Mgo().Collection("notifications").FindOne(context.Background(), bson.M{"_id": id}).Decode(&notification)
	return
}

// Fetch multiple leads by conditions
func FetchBy(deps Deps, query common.Query) (list Notifications, err error) {
	cursor, err := query(deps.Mgo().Collection("notifications"), context.Background())
	if err != nil {
		return
	}
	defer cursor.Close(context.Background())
	err = cursor.All(context.Background(), &list)
	return
}

func UserID(id primitive.ObjectID, take, skip int) common.Query {
	return func(col *mongo.Collection, ctx context.Context) (*mongo.Cursor, error) {
		opts := options.Find().SetLimit(int64(take)).SetSkip(int64(skip)).SetSort(bson.M{"updated_at": -1})
		return col.Find(ctx, bson.M{"user_id": id}, opts)
	}
}
