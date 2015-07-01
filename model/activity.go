package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Activity struct {
	Id  		bson.ObjectId `bson:"_id,omitempty" json:"id"`
	UserId 		bson.ObjectId `bson:"user_id,omitempty" json:"user_id"`
	Event 	  	string `bson:"event,omitempty" event:"related"`
	RelatedId 	bson.ObjectId `bson:"related_id,omitempty" json:"related_id,omitempty"`
	List    	[]bson.ObjectId `bson:"list,omitempty" json:"list,omitempty"`
	Created  	time.Time     `bson:"created_at" json:"created_at"`
}