package gcommerce

import (
	"errors"
)

// Set DI instance
func (this *Order) SetDI(di *Module) {

	this.di = di

	// Setup gateway
	gateway, err := getGateway(this.Gateway)

	if err != nil {
		panic(err)
	}

	this.gateway = gateway
	this.gateway.SetDI(this.di)
}

func (this *Order) Add(name, description, image string, price float64, q int, meta map[string]interface{}) {

	// Update price based on gateway
	gateway_price := this.gateway.ModifyPrice(price)

	this.Items = append(this.Items, Item{name, image, description, gateway_price, q, meta})
	this.Total = this.Total + gateway_price
}

func (this *Order) Ship(price float64, name string, address Address) {

	this.Shipping.Price = price
	this.Shipping.Type = name
	this.Shipping.Address = address

	this.Total = this.Total + price
}

func (this *Order) Checkout() error {

	database := this.di.Mongo.Database
	err := database.C("gcommerce_orders").Insert(this)

	if err != nil {
		return errors.New("internal-error")
	}

	return nil
}