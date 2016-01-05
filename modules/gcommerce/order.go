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
	this.gateway.SetOrder(this)
}

func (this *Order) Add(name, description, image string, price float64, q int, meta map[string]interface{}) {

	// Update price based on gateway
	gateway_price := this.gateway.ModifyPrice(price)

	this.Items = append(this.Items, Item{name, image, description, gateway_price, q, meta})
	this.Total = this.Total + gateway_price
}

func (this *Order) Ship(price float64, name string, address *CustomerAddress) {

	gateway_price := this.gateway.ModifyPrice(price)

	this.Shipping.Price = gateway_price
	this.Shipping.Type = name
	this.Shipping.Address = address.Address

	// Save the address reference in case we need it
	this.Shipping.Meta = map[string]interface{}{
		"related_id": address.Id,
	}

	// Use the address once
	address.UseOnce()

	this.Total = this.Total + gateway_price
}

func (this *Order) GetTotal() float64 {
	return this.gateway.AdjustPrice(this.Total)
}

func (this *Order) Checkout() error {

	// Global price mutators
	this.Total = this.gateway.AdjustPrice(this.Total)

	database := this.di.Mongo.Database

	// Perform the save of the order once we've got here
	err := database.C("gcommerce_orders").Insert(this)

	if err != nil {
		return errors.New("internal-error")
	}

	// Charge the user
	err = this.gateway.Charge(this.Total)

	return nil
}
