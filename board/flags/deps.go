package flags

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/siddontang/ledisdb/ledis"
	"gopkg.in/mgo.v2"
)

type deps interface {
	Mgo() *mgo.Database
	S3() *s3.Bucket
	LedisDB() *ledis.DB
}
