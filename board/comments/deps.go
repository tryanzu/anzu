package comments

import (
	"github.com/tidwall/buntdb"
	"github.com/tryanzu/core/modules/mail"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	BuntDB() *buntdb.DB
	Mgo() *mgo.Database
	Mailer() mail.Mailer
}
