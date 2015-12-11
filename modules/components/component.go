package components

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
	"math"
)

// Set DI instance
func (component *ComponentModel) SetDI(di *Module) {
	component.di = di
}

// Set generic data for the component model
func (component *ComponentModel) SetGeneric(data []byte) {
	component.generic = data
}

// Get generic data
func (component *ComponentModel) GetData() map[string]interface{} {

	var data map[string]interface{}

	err := bson.Unmarshal(component.generic, &data)

	if err != nil {
		panic(err)
	}

	return data
}

// Get Store price for vendor
func (component *ComponentModel) GetVendorPrice(vendor string) (float64, error) {

	if price, exists := component.Store.Prices[vendor]; exists {

		return price, nil
	}

	return float64(math.NaN()), exceptions.NotFound{"Invalid vendor. Not found."}
}

// Update component's price 
func (component *ComponentModel) UpdatePrice(prices map[string]float64) {

	database := component.di.Mongo.Database

	// Record price history
	if len(component.Store.Prices) > 0 {

		historic := &ComponentHistoricModel{
			ComponentId: component.Id,
			Store: component.Store,
			Created: time.Now(),
		}

		err := database.C("components_historic").Insert(historic)

		if err != nil {
			panic(err)
		}
	}


	set := bson.M{"store.updated_at": time.Now(), "activated": true}

	for key, price := range prices {

		set["store.prices." + key] = price
	}

	err := database.C("components").Update(bson.M{"_id": component.Id}, bson.M{"$set": set})

	if err != nil {
		panic(err)
	}

	go component.UpdateAlgolia()
}

// Perform component's algolia's record sync
func (component *ComponentModel) UpdateAlgolia() {

	index := component.di.Search.Get("components")

	// Compose algolia item
	full_name := component.FullName 

	if full_name == "" {
		full_name = component.Name
	}

	var image string

	if len(component.Images) > 0 {

		image = component.Images[0].Path
		image = strings.Replace(image, "full/", "", -1)
	}

	object := make(map[string]interface{})
	object["objectID"] = component.Id.Hex()
	object["name"] = component.Name
	object["full_name"] = full_name
	object["part_number"] = component.PartNumber
	object["slug"] = component.Slug
	object["image"] = image
	object["type"] = component.Type
	object["activated"] = true

	price, with_price := component.GetVendorPrice("spartangeek")

	if with_price == nil {

		object["price"] = price
	}

	_, err := index.UpdateObject(object)

	if err != nil {
		panic(err)
	}	
}