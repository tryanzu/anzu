package common

import (
	"gopkg.in/mgo.v2/bson"
)

func ById(id bson.ObjectId) bson.M {
	return bson.M{"_id": id}
}
