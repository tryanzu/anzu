package notifications

import (
	"context"
	"time"

	"github.com/tryanzu/core/deps"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func databaseWorker(n int) {
	for n := range Database {
		n.Id = primitive.NewObjectID()
		n.Seen = false
		n.Created = time.Now()
		n.Updated = time.Now()

		_, err := deps.Container.Mgo().Collection("notifications").InsertOne(context.Background(), n)
		if err != nil {
			panic(err)
		}

		_, err = deps.Container.Mgo().Collection("users").UpdateOne(
			context.Background(),
			bson.M{"_id": n.UserId},
			bson.M{"$inc": bson.M{"notifications": 1}},
		)
		if err != nil {
			panic(err)
		}

		var u struct {
			Count int `bson:"notifications"`
		}

		opts := options.FindOne().SetProjection(bson.M{"notifications": 1})
		err = deps.Container.Mgo().Collection("users").FindOne(context.Background(), bson.M{"_id": n.UserId}, opts).Decode(&u)
		if err != nil {
			panic(err)
		}

		Transmit <- Socket{"user " + n.UserId.Hex(), "notification", map[string]interface{}{
			"fire":  "notification",
			"count": u.Count,
		}}
	}
}
