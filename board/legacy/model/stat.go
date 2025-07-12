package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StatsComments struct {
	Id    primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Count int                `bson:"count" json:"count"`
}

type Stats struct {
	Comments int `json:"comments"`
	Users    int `json:"users"`
	Posts    int `json:"posts"`
}
