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
