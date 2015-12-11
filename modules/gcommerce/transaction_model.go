package gcommerce

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Transaction struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	OrderId   bson.ObjectId `bson:"order_id" json:"order_id"`
	Gateway   string        `bson:"gateway" json:"gateway"`
	Response  interface{}   `bson:"response" json:"response"`
	Error     interface{}   `bson:"error,omitempty" json:"error,omitempty"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
	Updated   time.Time     `bson:"updated_at" json:"updated_at"`
}