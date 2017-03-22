package deps

import (
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/olebedev/config"
	"github.com/op/go-logging"
	"gopkg.in/mgo.v2"
)

type Deps struct {
	ConfigProvider          *config.Config
	DatabaseSessionProvider *mgo.Session
	DatabaseProvider        *mgo.Database
	LoggerProvider          *logging.Logger
	MailerProvider          *mail.Module
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

func (d Deps) Mailer() *mail.Module {
	return d.MailerProvider
}
