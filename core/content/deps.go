package content

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/tryanzu/core/core/config"
	"go.mongodb.org/mongo-driver/mongo"
)

type deps interface {
	Mgo() *mongo.Database
	S3() *s3.Bucket
	LedisDB() *ledis.DB
}

func Boot() {
	log.SetBackend(config.LoggingBackend)
	go func() {
		for {
			<-config.C.Reload
			log.SetBackend(config.LoggingBackend)
		}
	}()
}
