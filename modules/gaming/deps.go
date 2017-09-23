package gaming

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2"
)

type Deps interface {
	Mgo() *mgo.Database
	Cache() *goredis.Redis
	GamingConfig() *model.GamingRules
}
