package deps

import (
	"github.com/go-redis/redis/v8"
	"github.com/op/go-logging"
	"github.com/siddontang/ledisdb/ledis"
	"github.com/tryanzu/core/board/legacy/model"
	"go.mongodb.org/mongo-driver/mongo"
)

type Deps struct {
	GamingConfigProvider    *model.GamingRules
	DatabaseSessionProvider *mongo.Client
	DatabaseProvider        *mongo.Database
	LoggerProvider          *logging.Logger
	CacheProvider           *redis.Client
	S3Provider              *S3Service
	LedisProvider           *ledis.DB
}

func (d Deps) GamingConfig() *model.GamingRules {
	return d.GamingConfigProvider
}

func (d Deps) Log() *logging.Logger {
	return d.LoggerProvider

}

func (d Deps) Mgo() *mongo.Database {
	return d.DatabaseProvider
}

func (d Deps) LedisDB() *ledis.DB {
	return d.LedisProvider
}

func (d Deps) MgoSession() *mongo.Client {
	return d.DatabaseSessionProvider
}

func (d Deps) S3() *S3Service {
	return d.S3Provider
}
