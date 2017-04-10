package user

import (
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
}
