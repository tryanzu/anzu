package notifications

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Notification struct {
	Id        bson.ObjectId   `bson:"_id,omitempty" json:"id,omitempty"`
	UserId    bson.ObjectId   `bson:"user_id" json:"user_id"`
	Type      string          `bson:"type" json:"type"`
	RelatedId bson.ObjectId   `bson:"related_id" json:"related_id"`
	Users     []bson.ObjectId `bson:"users" json:"users"`
	Seen      bool            `bson:"seen" json:"seen"`
	Created   time.Time       `bson:"created_at" json:"created_at"`
	Updated   time.Time       `bson:"updated_at" json:"updated_at"`
}

type Socket struct {
	Chan   string
	Action string
	Params map[string]interface{}
}
