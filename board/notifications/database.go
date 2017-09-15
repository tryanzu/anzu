package notifications

import (
	"github.com/fernandez14/spartangeek-blacker/deps"
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
	}
}
