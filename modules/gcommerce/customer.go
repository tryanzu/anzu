package gcommerce

import (
	"reflect"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Set DI instance
func (this *Customer) SetDI(di *Module) {
	this.di = di
}

func (this *Customer) Address(country, state, city, postal_code, line1, line2, extra string) Address {

	address := Address{
		Country: country,
		State: state,
		City: city,
		PostalCode: postal_code,
		Line1: line1,
		Line2: line2,
		Extra: extra,
	}

	for _, a := range this.Addresses {

		if reflect.DeepEqual(a, address) {
			return a
		}
	}

	database := this.di.Mongo.Database
	err := database.C("customers").Update(bson.M{"_id": this.Id}, bson.M{"$push": bson.M{"addresses": address}, "$set": bson.M{"updated_at": time.Now()}})

	if err != nil {
		panic(err)
	}

	return address
}

func (this *Customer) NewOrder(gateway_name string, address Address, meta map[string]interface{}) (*Order, error) {

	order := &Order{
		Id: bson.NewObjectId(),
		Status: ORDER_AWAITING,
		Statuses: make([]Status, 0),
		UserId: this.Id,
		Items: make([]Items, 0),
		Total: 0,
		Gateway: gateway_name,
		Meta: meta,
		Created: time.Now(),
		Updated: time.Now(),
	}

	order.SetDI(this.di) 

	return order, nil
}