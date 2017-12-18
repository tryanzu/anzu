package notifications

import (
	"github.com/op/go-logging"
	"github.com/tidwall/buntdb"
	"github.com/tryanzu/core/modules/mail"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	Log() *logging.Logger
	Mailer() mail.Mailer
	BuntDB() *buntdb.DB
}
