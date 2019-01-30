package votes

import (
	"github.com/siddontang/ledisdb/ledis"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Deps interface {
	LedisDB() *ledis.DB
	Mgo() *mgo.Database
}

type Votable interface {
	VotableType() string
	VotableID() bson.ObjectId
}
