package votes

import (
	"github.com/tidwall/buntdb"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Deps interface {
	BuntDB() *buntdb.DB
	Mgo() *mgo.Database
}

type Votable interface {
	VotableType() string
	VotableID() bson.ObjectId
}
