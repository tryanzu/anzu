package post

import (
	"github.com/op/go-logging"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/tryanzu/core/core/config"
	"go.mongodb.org/mongo-driver/mongo"
)

type deps interface {
	Mgo() *mongo.Database
	LedisDB() *ledis.DB
}

var log = logging.MustGetLogger("search")

func Boot() {
	log.SetBackend(config.LoggingBackend)
	go func() {
		for {
			<-config.C.Reload
			log.SetBackend(config.LoggingBackend)
		}
	}()
}
