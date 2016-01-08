package gcommerce

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Set DI instance
func (this *CustomerAddress) SetDI(di *Module) {
	this.di = di
}

// Increases the number of times the address has been used
func (this *CustomerAddress) UseOnce() {

	database := this.di.Mongo.Database
	err := database.C("customer_addresses").Update(bson.M{"_id": this.Id}, bson.M{"$set": bson.M{"last_used": time.Now()}, "$inc": bson.M{"times_used": 1}})

	if err != nil {
		panic(err)
	}

	// Update in-memory data
	this.LastUsed = time.Now()
	this.TimesUsed = this.TimesUsed + 1
}

// Compute First address line
func (this *CustomerAddress) Line1() string {

	return this.Address.Line1 + ", " + this.Address.Line2
}

// Compute Second address line
func (this *CustomerAddress) Line2() string {

	return this.Address.Neighborhood + ", " + this.Address.City + ", " + this.Address.State + " " + this.Address.PostalCode
}

// Compute Extra address line
func (this *CustomerAddress) Extra() string {

	return this.Address.Extra
}
