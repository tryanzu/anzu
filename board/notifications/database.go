package notifications

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

func databaseWorker(n int, d Deps) {
	for n := range Database {
		n.Id = bson.NewObjectId()
		n.Seen = false
		n.Created = time.Now()
		n.Updated = time.Now()

		err := d.Mgo().C("notifications").Insert(n)
		if err != nil {
			d.Log().Error(err)
		}
	}
}
