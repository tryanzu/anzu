package assets

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/siddontang/ledisdb/ledis"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	S3() *s3.Bucket
	Config() *config.Config
	LedisDB() *ledis.DB
}
