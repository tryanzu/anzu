package model

import (
	"gopkg.in/mgo.v2/bson"
)

type StatsComments struct {
	Id  	bson.ObjectId `bson:"_id,omitempty" json:"_id,omitempty"`
	Count 	int       	  `bson:"count" json:"count"`
}

type Stats struct {
	Comments 	int `json:"comments"`
	Users 		int `json:"users"`
	Posts 		int `json:"posts"`
}