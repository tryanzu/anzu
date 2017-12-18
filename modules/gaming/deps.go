package gaming

import (
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	Cache() *goredis.Redis
	GamingConfig() *model.GamingRules
}
