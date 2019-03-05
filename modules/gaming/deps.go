package gaming

import (
	"github.com/tryanzu/core/board/legacy/model"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	GamingConfig() *model.GamingRules
}
