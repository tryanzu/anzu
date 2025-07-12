package common

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ById(id primitive.ObjectID) bson.M {
	return bson.M{"_id": id}
}
