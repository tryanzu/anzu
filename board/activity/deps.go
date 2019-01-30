package activity

import (
	"gopkg.in/mgo.v2"
)

type deps interface {
	Mgo() *mgo.Database
}
