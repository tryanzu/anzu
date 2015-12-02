package gcommerce

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func Boot() *Module {

	module := &Module{}

	return module
}

type Module struct {
	Mongo  *mongo.Service               `inject:""`
	Errors *exceptions.ExceptionsModule `inject:""`
	Redis  *goredis.Redis               `inject:""`
	Mail   *mail.Module                 `inject:""`
}

func (module *Module) GetCustomerFromUser(user_id bson.ObjectId) Customer {

	var customer Customer

	database := module.Mongo.Database
	err := database.C("customers").Find(bson.M{"user_id": user_id}).One(&customer)

	if err != nil {
		
		customer = Customer{
			Id: bson.NewObjectId(),
			UserId: user_id,
			Created: time.Now(),
			Updated: time.Now(),
		}

		err := database.C("customers").Insert(customer)

		if err != nil {
			panic(err)
		}
	} 

	return customer
}