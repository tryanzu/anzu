package activity

import (
	"github.com/op/go-logging"
	"gopkg.in/mgo.v2"
)

var (
	log = logging.MustGetLogger("activity")
)

type deps interface {
	Mgo() *mgo.Database
}
