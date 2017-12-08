package notifications

import (
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/transmit"
	"github.com/op/go-logging"
	"github.com/tidwall/buntdb"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	Transmit() transmit.Sender
	Log() *logging.Logger
	Mailer() mail.Mailer
	BuntDB() *buntdb.DB
}
