package users

import (
	"github.com/siddontang/ledisdb/ledis"
	"github.com/tryanzu/core/modules/mail"
	"gopkg.in/mgo.v2"
)

type deps interface {
	Mgo() *mgo.Database
	Mailer() mail.Mailer
	LedisDB() *ledis.DB
}
