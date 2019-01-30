package notifications

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/op/go-logging"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/tryanzu/core/modules/mail"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	Log() *logging.Logger
	Mailer() mail.Mailer
	S3() *s3.Bucket
	Config() *config.Config
	LedisDB() *ledis.DB
}
