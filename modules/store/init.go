package store 

import (
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/xuyu/goredis"
	"time"
)

type Module struct {
	Mongo   *mongo.Service               `inject:""`
	Errors  *exceptions.ExceptionsModule `inject:""`
	Redis   *goredis.Redis               `inject:""`
}

func (module *Module) CreateOrder(order OrderModel) {

	database := module.Mongo.Database

	// Set the dates
	order.Created = time.Now()
	order.Updated = time.Now()

	err := database.C("stats").Insert(order)

	if err != nil {
		panic(err)
	}
}