package security 

import (
	"gopkg.in/mgo.v2/bson"
)

type IpAddress struct {
	Id         bson.ObjectId   `bson:"_id,omitempty" json:"id,omitempty"`
	Address    string          `bson:"address" json:"address"`
	Users      []bson.ObjectId `bson:"users" json:"users"`
	Banned     bool            `bson:"banned" json:"banned"`
}