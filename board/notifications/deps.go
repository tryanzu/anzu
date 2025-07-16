package notifications

import (
	"github.com/siddontang/ledisdb/ledis"
	"github.com/tryanzu/core/deps"
	"go.mongodb.org/mongo-driver/mongo"
)

type Deps interface {
	Mgo() *mongo.Database
	S3() *deps.S3Service
	LedisDB() *ledis.DB
}
