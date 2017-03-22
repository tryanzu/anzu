package deps

import (
	"github.com/olebedev/config"
	"github.com/op/go-logging"
	"gopkg.in/mgo.v2"
)

type Deps struct {
	Config   *config.Config
	Session  *mgo.Session
	Database *mgo.Database
	Logger   *logging.Logger
}

func (d Deps) ConfigProvider() *config.Config {
	return d.Config
}

func (d Deps) LogProvider() *logging.Logger {
	return d.Logger
}

func (d Deps) MgoProvider() *mgo.Database {
	return d.Database
}
