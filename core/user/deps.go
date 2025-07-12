package user

import (
	"github.com/siddontang/ledisdb/ledis"
	"go.mongodb.org/mongo-driver/mongo"
)

type deps interface {
	Mgo() *mongo.Database
	LedisDB() *ledis.DB
}
