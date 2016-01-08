package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

type CartAPI struct {
	Components *components.Module `inject:""`
}

// Get Cart items
func (this CartAPI) Get(c *gin.Context) {

	var items []CartComponentItem

	// Initialize cart library
	err := this.getCart(c).Bind(&items)

	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
	} else {

		if len(items) > 0 {
			c.JSON(200, items)
		} else {
			c.JSON(200, make([]string, 0))
		}
	}
}

// Add Cart item from component id
func (this CartAPI) Add(c *gin.Context) {

	var form CartAddForm

	if c.BindJSON(&form) == nil {

		id := form.Id

		if !bson.IsObjectIdHex(id) {
			c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
			return
		}

		// Get the component and its data
		component_id := bson.ObjectIdHex(id)
		component, err := this.Components.Get(component_id)

		if err != nil {
			c.JSON(404, gin.H{"message": "Invalid request, component not found.", "status": "error"})
			return
		}

		// Initialize cart library
		container := this.getCart(c)
		{
			var items []CartComponentItem

			err := container.Bind(&items)

			if err != nil {
				c.JSON(500, gin.H{"status": "error", "message": err.Error()})
				return
			}

			price, err := component.GetVendorPrice(form.Vendor)

			if err != nil {
				c.JSON(400, gin.H{"message": "Invalid vendor, check id.", "status": "error"})
				return
			}

			attrs := map[string]interface{}{
				"vendor": form.Vendor,
			}

			exists := false

			for index, item := range items {

				if item.Id == component.Id.Hex() {

					exists = true

					items[index].IncQuantity(1) 
					break
				}
			}

			if !exists {

				base := cart.CartItem{
					Id: component.Id.Hex(),
					Name: component.Name,
					Price: price,
					Quantity: 1,
					Attributes: attrs,
				}

				item := CartComponentItem{base, component.FullName, component.Image, component.Slug, component.Type}
				items = append(items, item)
			}
			
			err = container.Save(items)

			if err != nil {
				c.JSON(500, gin.H{"status": "error", "message": err.Error()})
				return
			}
		}

		c.JSON(200, gin.H{"status": "okay"})
	}
}

// Delete Item from cart
func (this CartAPI) Delete(c *gin.Context) {

	id := c.Param("id")

	if !bson.IsObjectIdHex(id) {
		c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
		return
	}

	container := this.getCart(c)
	{
		var items []CartComponentItem

		err := container.Bind(&items)

		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": err.Error()})
			return
		}

		for i, item := range items {

			if item.Id == id {

				items = append(items[:i], items[i+1:]...)
				break
			}
		}

		container.Save(items)

		c.JSON(200, gin.H{"status": "okay"})
	}
}

func (this CartAPI) getCart(c *gin.Context) *cart.Cart {

	obj, err := cart.Boot(cart.GinGonicSession{sessions.Default(c)})

	if err != nil {
		panic(err)
	}

	return obj
}

type CartAddForm struct {
	Id     string `json:"id" binding:"required"`
	Vendor string `json:"vendor" binding:"required"`
}

type CartComponentItem struct {
	cart.CartItem
	FullName string `json:"full_name"`
	Image    string `json:"image"`
	Slug     string `json:"slug"`
	Type     string `json:"type"` 
}