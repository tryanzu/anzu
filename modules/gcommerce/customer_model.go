package gcommerce

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Customer struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	UserId    bson.ObjectId `bson:"user_id" json:"user_id"`
	Addresses []Address     `bson:"addresses" json:"addresses"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
	Updated   time.Time     `bson:"updated_at" json:"updated_at"`
}