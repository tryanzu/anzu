package votes

import (
	"gopkg.in/mgo.v2/bson"

	"time"
)

type Vote struct {
	Id         bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	UserId     bson.ObjectId `bson:"user_id" json:"user_id"`
	Type       string        `bson:"type" json:"type"`
	NestedType string        `bson:"nested_type,omitempty" json:"nested_type,omitempty"`
	RelatedId  bson.ObjectId `bson:"related_id" json:"related_id"`
	Value      int           `bson:"value" json:"value"`
	Created    time.Time     `bson:"created_at" json:"created_at"`
}

type Votes struct {
	Up     int `bson:"up" json:"up"`
	Down   int `bson:"down" json:"down"`
	Rating int `bson:"rating,omitempty" json:"rating,omitempty"`
}
