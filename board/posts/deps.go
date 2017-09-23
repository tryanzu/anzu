package post

import (
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
}
