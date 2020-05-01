package notifications

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/op/go-logging"
	"github.com/siddontang/ledisdb/ledis"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	Log() *logging.Logger
	S3() *s3.Bucket
	LedisDB() *ledis.DB
}
