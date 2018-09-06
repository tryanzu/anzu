package common

import (
	"gopkg.in/mgo.v2/bson"
)

type UsersStringMap map[bson.ObjectId]string
type AssetsStringMap map[bson.ObjectId]string
