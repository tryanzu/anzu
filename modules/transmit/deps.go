package transmit

import (
	"github.com/olebedev/config"
	"github.com/op/go-logging"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	Cache() *goredis.Redis
	Config() *config.Config
	Log() *logging.Logger
}
