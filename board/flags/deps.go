package flags

import (
	"github.com/siddontang/ledisdb/ledis"
	"github.com/tryanzu/core/deps"
	"go.mongodb.org/mongo-driver/mongo"
)

type DepsInterface interface {
	Mgo() *mongo.Database
	S3() *deps.S3Service
	LedisDB() *ledis.DB
}
