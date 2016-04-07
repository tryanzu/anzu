package gcommerce

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
)

// Initialize Product model struct (single one)
func (this *Product) Initialize() {
	
	// Load component if needed
	if this.Type == "component" {

		component_id, exists := this.Attrs["component_id"].(bson.ObjectId)

		if exists {
			component, err := this.di.Components.Get(component_id)

			if err != nil {
				panic(err)
			}

			this.ComponentBind(component)
		}
	}
}

// Use Component pointer to fulfill product fields
func (this *Product) ComponentBind(component *components.ComponentModel) {

	this.Name = component.FullName
	this.Description = component.Name
	this.Slug = component.Slug
	this.Category = component.Type
	this.Image = component.Image
	this.Images = component.Images
	this.Attrs = component.GetData()

	price, err := component.GetVendorPrice("spartangeek")

	if err != nil {
		panic(err)
	}

	this.Price = price
}

// Initialize Product models list (many)
func (this Products) InitializeList(list []*Product) {

	var component_ids []bson.ObjectId

	for _, product := range list {

		if product.Type == "component" {

			component_id, exists := product.Attrs["component_id"].(bson.ObjectId)

			if exists {
				component_ids = append(component_ids, component_id)
			}
		}
	}

	// Component products eager loading
	if len(component_ids) > 0 {

		var primitives []interface{}
		var components map[string]*components.ComponentModel

		database := this.di.Mongo.Database
		err := database.C("components").Find(bson.M{"_id": bson.M{"$in": component_ids}}).All(&primitives)

		if err != nil {
			panic(err)
		}

		// Use primitives to generate components map
		for _, component := range primitives {


			// Marshal the data inside the generic model
			encoded, err := bson.Marshal(component)

			if err != nil {
				panic(err)
			}

			c, err := this.di.Components.Get(encoded)

			if err != nil {
				panic(err)
			}

			components[c.Id.Hex()] = c
		}

		for index, product := range list {

			if product.Type == "component" {

				if component_id, exists := product.Attrs["component_id"].(bson.ObjectId); exists {
					
					if component, ref_exists := components[component_id.Hex()]; ref_exists {

						list[index].ComponentBind(component)
					}
				}
			}
		}
	}
}