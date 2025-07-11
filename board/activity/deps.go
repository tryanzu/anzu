package activity

import (
	"github.com/op/go-logging"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	log = logging.MustGetLogger("activity")
)

type deps interface {
	Mgo() *mongo.Database
}
