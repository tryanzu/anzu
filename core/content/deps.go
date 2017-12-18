package content

import (
	"github.com/tidwall/buntdb"
	"github.com/tryanzu/core/modules/mail"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	Mailer() mail.Mailer
	BuntDB() *buntdb.DB
}
