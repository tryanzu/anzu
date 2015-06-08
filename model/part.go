package model

import (
	"gopkg.in/mgo.v2/bson"
)

type PartByModel struct {
	Id 		bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	Name 	string                 `bson:"name" json:"name"`
}