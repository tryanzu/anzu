package assets

import (
	"github.com/tidwall/buntdb"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	BuntDB() *buntdb.DB
}
