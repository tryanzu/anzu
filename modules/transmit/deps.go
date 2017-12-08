package transmit

import (
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/olebedev/config"
	"github.com/op/go-logging"
	"github.com/tidwall/buntdb"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	Cache() *goredis.Redis
	Config() *config.Config
	Log() *logging.Logger
	Mailer() mail.Mailer
	BuntDB() *buntdb.DB
}
