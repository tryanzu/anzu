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
	search := c.Query("search")

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

	meta := bson.M{}

	if search != "" {

		meta = bson.M{
			"$or": []bson.M{
				{
					"reference": bson.M{
						"$regex": search,
					},
				},
				{
					"$text": bson.M{
						"$search": search,
					},
				},
			},
		}
	}

	orders := this.GCommerce.Get(meta, limit, offset)

	c.JSON(200, orders)
}