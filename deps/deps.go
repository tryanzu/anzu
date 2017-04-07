package deps

import (
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/olebedev/config"
	"github.com/op/go-logging"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2"
)

type Deps struct {
	ConfigProvider          *config.Config
	DatabaseSessionProvider *mgo.Session
	DatabaseProvider        *mgo.Database
	LoggerProvider          *logging.Logger
	MailerProvider          mail.Mailer
	CacheProvider           *goredis.Redis
}

func (d Deps) Config() *config.Config {
	return d.ConfigProvider
}

func (d Deps) Log() *logging.Logger {
	return d.LoggerProvider
}

func (d Deps) Mgo() *mgo.Database {
	return d.DatabaseProvider
}

func (d Deps) Mailer() mail.Mailer {
	return d.MailerProvider
}

func (d Deps) Cache() *goredis.Redis {
	return d.CacheProvider
}
