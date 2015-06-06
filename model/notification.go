package model

import (
	"time"
	"gopkg.in/mgo.v2/bson"
)

type Counter struct {
    UserId   bson.ObjectId `bson:"user_id" json:"user_id"`
    Counters map[string]PostCounter `bson:"counters" json:"counters"`
}

type PostCounter struct {
    Counter int `bson:"counter" json:"counter"`
    Updated time.Time `bson:"updated_at" json:"updated_at"`
}

type Notification struct {
    Id        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
    UserId    bson.ObjectId `bson:"user_id" json:"user_id"`
    RelatedId bson.ObjectId `bson:"related_id" json:"related_id"`
    Title     string `bson:"title" json:"title"`
    Text      string `bson:"text" json:"text"`
    Link      string `bson:"link" json:"link"`
    Related   string `bson:"related" json:"related"`
    Seen      bool   `bson:"seen" json:"seen"`
    Image     string `bson:"image" json:"image"`
    Created   time.Time `bson:"created_at" json:"created_at"`
    Updated   time.Time `bson:"updated_at" json:"updated_at"`
}