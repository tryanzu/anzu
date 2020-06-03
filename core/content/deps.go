package content

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/tryanzu/core/core/config"
	"gopkg.in/mgo.v2"
)

type deps interface {
	Mgo() *mgo.Database
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
