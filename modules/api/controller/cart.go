package controller 

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/gin-gonic/gin"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/gin-gonic/contrib/sessions"
)

type CartAPI struct {
	Components *components.Module `inject:""`
}

// Get Cart items
func (this CartAPI) Get(c *gin.Context) {

	// Initialize cart library
	container := this.getCart(c)

	c.JSON(200, container.GetContent())
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
			price, err := component.GetVendorPrice(form.Vendor)

			if err != nil {
				c.JSON(400, gin.H{"message": "Invalid vendor, check id.", "status": "error"})
				return
			}

			attrs :=  map[string]interface{}{
				"vendor": form.Vendor,
			}

			container.Add(component.Id.Hex(), component.Name, price, 1, attrs)
		}

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