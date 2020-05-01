package post

import (
	"github.com/op/go-logging"
	"github.com/siddontang/ledisdb/ledis"
	"gopkg.in/mgo.v2"
)

type deps interface {
	Mgo() *mgo.Database
	LedisDB() *ledis.DB
}

var log = logging.MustGetLogger("search")
