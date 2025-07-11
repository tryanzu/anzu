package security 

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IpAddress struct {
	Id         primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Address    string               `bson:"address" json:"address"`
	Users      []primitive.ObjectID `bson:"users" json:"users"`
	Banned     bool                 `bson:"banned" json:"banned"`
}