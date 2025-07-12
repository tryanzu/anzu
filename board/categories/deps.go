package categories

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type deps interface {
	Mgo() *mongo.Database
}
