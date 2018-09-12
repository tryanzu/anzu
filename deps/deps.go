package deps

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/op/go-logging"
	"github.com/tidwall/buntdb"
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/modules/mail"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2"
)

type Deps struct {
	ConfigProvider          *config.Config
	GamingConfigProvider    *model.GamingRules
	DatabaseSessionProvider *mgo.Session
	DatabaseProvider        *mgo.Database
	LoggerProvider          *logging.Logger
	MailerProvider          mail.Mailer
	CacheProvider           *goredis.Redis
	BuntProvider            *buntdb.DB
	S3Provider              *s3.Bucket
}

func (d Deps) Config() *config.Config {
	return d.ConfigProvider
}

func (d Deps) GamingConfig() *model.GamingRules {
	return d.GamingConfigProvider
}

func (d Deps) Log() *logging.Logger {
	return d.LoggerProvider

}

func (d Deps) Mgo() *mgo.Database {
	return d.DatabaseProvider
}

func (d Deps) MgoSession() *mgo.Session {
	return d.DatabaseSessionProvider
}

func (d Deps) BuntDB() *buntdb.DB {
	return d.BuntProvider
}

func (d Deps) S3() *s3.Bucket {
	return d.S3Provider
}

func (d Deps) Mailer() mail.Mailer {
	return d.MailerProvider
}

func (d Deps) Cache() *goredis.Redis {
	return d.CacheProvider
}
