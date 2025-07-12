package gaming

import (
	"github.com/tryanzu/core/board/legacy/model"
	"go.mongodb.org/mongo-driver/mongo"
)

type Deps interface {
	Mgo() *mongo.Database
	GamingConfig() *model.GamingRules
}
