package gcommerce

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/fernandez14/spartangeek-blacker/modules/components"

	"sort"
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

func (this *Product) InitializeMassdrop() {

	var model *Massdrop
	database := this.di.Mongo.Database


	err := database.C("gcommerce_massdrop").Find(bson.M{"product_id": this.Id}).One(&model)

	if err == nil {

		this.Massdrop = model

		var transactions []MassdropTransaction
		var activities []MassdropActivity

		err := database.C("gcommerce_massdrop_transactions").Find(bson.M{"massdrop_id": model.Id}).Sort("-created_at").All(&transactions)

		if err != nil {
			panic(err)
		}

		if len(transactions) > 0 {

			var customers []bson.ObjectId

			for _, t := range transactions {
				customers = append(customers, t.CustomerId)
			}

			customers_map, users_map := this.di.JoinUsers(customers)

			var reservations int = 0
			var interested int = 0

			for _, t := range transactions {

				if t.Status == MASSDROP_STATUS_COMPLETED {

					// User information
					customer := customers_map[t.CustomerId]
					usr := users_map[customer.UserId]

					activity := MassdropActivity{
						Type: t.Type,
						Created: t.Created,
						Attrs: map[string]interface{}{
							"user": usr,
						},
					}

					activities = append(activities, activity)

					if t.Type == MASSDROP_TRANS_RESERVATION {
						reservations = reservations + 1
					} else if t.Type == MASSDROP_TRANS_INSTERESTED {
						interested = interested + 1
					}
				} 
			}

			// First activities sorting
			sort.Sort(MassdropByCreated(activities))

			for index, c := range this.Massdrop.Checkpoints {

				if reservations >= c.Ends  {

					this.Massdrop.Checkpoints[index].Done = true

					count := 0

					for _, act := range activities {

						if act.Type != MASSDROP_TRANS_RESERVATION {
							continue
						} 

						count = count + 1

						if count == c.Ends {

							activity := MassdropActivity{
								Type: "checkpoint",
								Created: act.Created,
								Attrs: map[string]interface{}{
									"step": c.Step,
								},
							}

							activities = append(activities, activity)
							break
						}
					}
				}
			}

			sort.Sort(MassdropByCreated(activities))

			this.Massdrop.Activities = activities
			this.Massdrop.Reservations = reservations
			this.Massdrop.Interested = interested
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
		components_map := make(map[string]*components.ComponentModel)

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

			var model *components.ComponentModel

			err = bson.Unmarshal(encoded, &model)

			if err != nil {
				panic(err)
			}

			model.SetGeneric(encoded)

			c, err := this.di.Components.Get(model)

			if err != nil {
				panic(err)
			}

			components_map[c.Id.Hex()] = c
		}

		for index, product := range list {

			if product.Type == "component" {

				if component_id, exists := product.Attrs["component_id"].(bson.ObjectId); exists {
					
					if component, ref_exists := components_map[component_id.Hex()]; ref_exists {

						list[index].ComponentBind(component)
					}
				}
			}
		}
	}
}