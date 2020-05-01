package user

import (
	"github.com/siddontang/ledisdb/ledis"
	"gopkg.in/mgo.v2"
)

type deps interface {
	Mgo() *mgo.Database
	LedisDB() *ledis.DB
}
