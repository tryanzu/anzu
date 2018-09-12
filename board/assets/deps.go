package assets

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/tidwall/buntdb"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	BuntDB() *buntdb.DB
	S3() *s3.Bucket
	Config() *config.Config
}
