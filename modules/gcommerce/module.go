package gcommerce

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/olebedev/config"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"

	"time"
)

func Boot(key string) *Module {

	module := &Module{StripeKey: key}

	return module
}

func ValidStatus(name string) bool {
	
	if name == ORDER_CONFIRMED ||
	   name == ORDER_AWAITING ||
	   name == ORDER_INSTOCK ||
	   name == ORDER_SHIPPED ||
	   name == ORDER_COMPLETED ||
	   name == ORDER_CANCELED ||
	   name == ORDER_PAYMENT_ERROR {
	   	
	   	return true
	}
	   
	return false
}

type Module struct {
	Mongo     *mongo.Service               `inject:""`
	Errors    *exceptions.ExceptionsModule `inject:""`
	Redis     *goredis.Redis               `inject:""`
	Config    *config.Config               `inject:""`
	Mail      *mail.Module                 `inject:""`
	StripeKey string
}

func (module *Module) GetCustomerFromUser(user_id bson.ObjectId) Customer {

	var customer Customer

	database := module.Mongo.Database
	err := database.C("customers").Find(bson.M{"user_id": user_id}).One(&customer)

	if err != nil {

		customer = Customer{
			Id:      bson.NewObjectId(),
			UserId:  user_id,
			Created: time.Now(),
			Updated: time.Now(),
		}

		err := database.C("customers").Insert(customer)

		if err != nil {
			panic(err)
		}
	}

	customer.SetDI(module)

	return customer
}

func (module *Module) GetCustomer(id bson.ObjectId) (*Customer, error) {
	
	var customer *Customer

	database := module.Mongo.Database
	err := database.C("customers").Find(bson.M{"_id": id}).One(&customer)

	if err != nil {
		return nil, err
	}

	customer.SetDI(module)

	return customer, nil
}

func (module *Module) Get(where bson.M, limit, offset int) []Order {

	var list []Order
	var customers []Customer
	var users []user.UserBasic

	var customer_ids []bson.ObjectId
	var user_ids []bson.ObjectId

	database := module.Mongo.Database
	err := database.C("gcommerce_orders").Find(where).Limit(limit).Skip(offset).Sort("-created_at").All(&list)

	if err != nil {
		panic(err)
	}

	for _, order := range list {
		customer_ids = append(customer_ids, order.UserId)
	}

	err = database.C("customers").Find(bson.M{"_id": bson.M{"$in": customer_ids}}).All(&customers)

	if err != nil {
		panic(err)
	}

	for _, customer := range customers {
		user_ids = append(user_ids, customer.UserId)
	}

	err = database.C("users").Find(bson.M{"_id": bson.M{"$in": user_ids}}).Select(bson.M{"_id": 1, "username": 1, "username_slug": 1, "email": 1, "facebook": 1, "validated": 1, "banned": 1, "created_at": 1, "updated_at": 1}).All(&users)

	if err != nil {
		panic(err)
	}

	customer_map := map[bson.ObjectId]Customer{}

	for _, customer := range customers {
		customer_map[customer.Id] = customer
	}	

	users_map := map[bson.ObjectId]user.UserBasic{}

	for _, usr := range users {
		users_map[usr.Id] = usr
	}	

	for index, order := range list {

		c := customer_map[order.UserId]
		usr := users_map[c.UserId]

		list[index].Customer = c
		list[index].User = usr
	}

	return list
}

func (module *Module) One(where bson.M) (*Order, error) {

	var order *Order

	database := module.Mongo.Database
	err := database.C("gcommerce_orders").Find(where).One(&order)

	if err != nil {
		return nil, err
	}

	order.SetDI(module)

	return order, nil
}