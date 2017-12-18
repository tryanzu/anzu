package notifications

import (
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func databaseWorker(n int) {
	for n := range Database {
		n.Id = bson.NewObjectId()
		n.Seen = false
		n.Created = time.Now()
		n.Updated = time.Now()

		err := deps.Container.Mgo().C("notifications").Insert(n)
		if err != nil {
			panic(err)
		}

		err = deps.Container.Mgo().C("users").Update(bson.M{"_id": n.UserId}, bson.M{"$inc": bson.M{"notifications": 1}})
		if err != nil {
			panic(err)
		}

		var u struct {
			Count int `bson:"notifications"`
		}

		err = deps.Container.Mgo().C("users").FindId(n.UserId).Select(bson.M{"notifications": 1}).One(&u)
		if err != nil {
			panic(err)
		}

		Transmit <- Socket{"user " + n.UserId.Hex(), "notification", map[string]interface{}{
			"fire":  "notification",
			"count": u.Count,
		}}
	}
}
