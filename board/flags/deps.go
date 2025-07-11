package flags

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/siddontang/ledisdb/ledis"
	"go.mongodb.org/mongo-driver/mongo"
)

type deps interface {
	Mgo() *mongo.Database
	S3() *s3.Bucket
	LedisDB() *ledis.DB
}
