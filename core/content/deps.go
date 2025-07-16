package content

import (
	"github.com/siddontang/ledisdb/ledis"
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/deps"
	"go.mongodb.org/mongo-driver/mongo"
)

type DepsInterface interface {
	Mgo() *mongo.Database
	S3() *deps.S3Service
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
