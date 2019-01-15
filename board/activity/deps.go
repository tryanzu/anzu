package activity

import (
	"github.com/tidwall/buntdb"
	"gopkg.in/mgo.v2"
)

type deps interface {
	Mgo() *mgo.Database
	BuntDB() *buntdb.DB
}
