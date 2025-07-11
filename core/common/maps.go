package common

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UsersStringMap is just map of id -> username (cache purposes).
type UsersStringMap map[primitive.ObjectID]string

type AssetRef struct {
	URL         string
	UseOriginal bool
}

type AssetRefsMap map[primitive.ObjectID]AssetRef
