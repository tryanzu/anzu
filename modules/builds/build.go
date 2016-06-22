package builds

import (
	"gopkg.in/mgo.v2/bson"

	"time"
)

type Build struct {
	Id          bson.ObjectId   `bson:"_id,omitempty" json:"id,omitempty"`
	Ref         string          `bson:"ref" json:"ref"`
	UserId      *bson.ObjectId  `bson:"user_id,omitempty" json:"user_id"`
	SessionId   string          `bson:"session_id" json:"session_id"`
	Name        string          `bson:"name" json:"name"`
	CPU         bson.ObjectId   `bson:"cpu,omitempty" json:"-"`
	Cooler      bson.ObjectId   `bson:"cooler,omitempty" json:"-"`
	Motherboard bson.ObjectId   `bson:"motherboard,omitempty" json:"-"`
	GPU         []bson.ObjectId `bson:"gpu,omitempty" json:"-"`
	Memory      []bson.ObjectId `bson:"memory,omitempty" json:"-"`
	Storage     []bson.ObjectId `bson:"storage,omitempty" json:"-"`
	Case        bson.ObjectId   `bson:"case,omitempty" json:"-"`
	PSU         bson.ObjectId   `bson:"psu,omitempty" json:"-"`
	Created     time.Time       `bson:"created_at" json:"created_at"`
	Updated     time.Time       `bson:"updated_at" json:"updated_at"`
}
