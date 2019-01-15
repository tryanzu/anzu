package activity

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// M stands for activity model.
type M struct {
	ID        bson.ObjectId   `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectId   `bson:"user_id,omitempty" json:"user_id"`
	Event     string          `bson:"event,omitempty" event:"related"`
	RelatedID bson.ObjectId   `bson:"related_id,omitempty" json:"related_id,omitempty"`
	List      []bson.ObjectId `bson:"list,omitempty" json:"list,omitempty"`
	Created   time.Time       `bson:"created_at" json:"created_at"`
}
