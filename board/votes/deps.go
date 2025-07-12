package votes

import (
	"github.com/siddontang/ledisdb/ledis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Deps interface {
	LedisDB() *ledis.DB
	Mgo() *mongo.Database
}

type Votable interface {
	VotableType() string
	VotableID() primitive.ObjectID
}
