package content

import (
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/tidwall/buntdb"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	Mailer() mail.Mailer
	BuntDB() *buntdb.DB
}
