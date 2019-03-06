package flags

import (
	"html"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// UpsertComment performs validations before upserting data struct
func UpsertFlag(d deps, f Flag) (flag Flag, err error) {
	if f.ID.Valid() == false {
		f.ID = bson.NewObjectId()
		f.Created = time.Now()
		f.Status = PENDING
	}

	f.Content = html.EscapeString(f.Content)
	f.Updated = time.Now()
	changes, err := d.Mgo().C("flags").UpsertId(f.ID, bson.M{"$set": f})
	if err != nil {
		return
	}

	if changes.Matched == 0 {
		// When inserted
	}
	flag = f
	return
}
