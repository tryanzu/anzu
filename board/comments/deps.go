package comments

import (
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/tidwall/buntdb"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	BuntDB() *buntdb.DB
	Mgo() *mgo.Database
	Mailer() mail.Mailer
}
