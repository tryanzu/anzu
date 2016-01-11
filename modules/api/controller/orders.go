package controller 

import (
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

	"strconv"
)

type OrdersAPI struct {
	GCommerce  *gcommerce.Module  `inject:""`
}

func (this OrdersAPI) Get(c *gin.Context) {

	var offset int = 0
	var limit  int = 20

	l := c.Query("limit")
	o := c.Query("offset")

	if l != "" {
		cl, err := strconv.Atoi(l)

		if err == nil && cl > 0 {
			limit = cl 
		}
	}

	if o != "" {
		co, err := strconv.Atoi(o)

		if err == nil && co > 0 {
			offset = co 
		}
	}

	orders := this.GCommerce.Get(bson.M{}, limit, offset)

	c.JSON(200, orders)
}