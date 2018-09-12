package notifications

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/op/go-logging"
	"github.com/tidwall/buntdb"
	"github.com/tryanzu/core/modules/mail"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	Log() *logging.Logger
	Mailer() mail.Mailer
	BuntDB() *buntdb.DB
	S3() *s3.Bucket
	Config() *config.Config
}
