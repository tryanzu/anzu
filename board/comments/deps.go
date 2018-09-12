package comments

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/tidwall/buntdb"
	"github.com/tryanzu/core/modules/mail"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	BuntDB() *buntdb.DB
	Mgo() *mgo.Database
	Mailer() mail.Mailer
	S3() *s3.Bucket
	Config() *config.Config
}
