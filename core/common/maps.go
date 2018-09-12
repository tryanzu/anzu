package common

import (
	"gopkg.in/mgo.v2/bson"
)

// UsersStringMap is just map of id -> username (cache purposes).
type UsersStringMap map[bson.ObjectId]string

type AssetRef struct {
	URL         string
	UseOriginal bool
}

type AssetRefsMap map[bson.ObjectId]AssetRef
