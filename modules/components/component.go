package components

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

func (component *ComponentModel) SetDI(di *Module) {

	component.di = di
}

func (component *ComponentModel) UpdatePrice(prices map[string]float64) {

	database := component.di.Mongo.Database
	set := bson.M{"store.updated_at": time.Now(), "activated": true}

	for key, price := range prices {

		set["store.prices." + key] = price
	}

	err := database.C("components").Update(bson.M{"_id": component.Id}, bson.M{"$set": set})

	if err != nil {
		panic(err)
	}
}