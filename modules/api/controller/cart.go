package controller 

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/gin-gonic/gin"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
)

type CartAPI struct {
	Components *components.Module `inject:""`
	Cart       *cart.Cart         `inject:""`
}

// Add Cart item from component id
func (this CartAPI) Add(c *gin.Context) {

	id := c.Param("id")

	if !bson.IsObjectIdHex(id) {

		c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
		return
	}

	component_id := bson.ObjectIdHex(id)
	component, err := this.Components.Get(component_id)

	if err != nil {
		c.JSON(404, gin.H{"message": "Invalid request, component not found.", "status": "error"})
		return
	}



	c.JSON(200, component.GetData())
}