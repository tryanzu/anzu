package gcommerce

import (
	"gopkg.in/mgo.v2/bson"
)

func (this *Product) Initialize() {
	
	// Load component if needed
	if this.Type == "component" {

		component_id, exists := this.Attrs["component_id"].(bson.ObjectId)

		if exists {
			component, err := this.di.Components.Get(component_id)

			if err != nil {
				panic(err)
			}

			this.Name = component.FullName
			this.Description = component.Name
			this.Slug = component.Slug
			this.Category = component.Type
			this.Image = component.Image
			this.Price, err = component.GetVendorPrice("spartangeek")
			this.Attrs = component.GetData()

			if err != nil {
				panic(err)
			}
		}
	}
}