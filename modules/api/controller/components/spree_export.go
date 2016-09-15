package components

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) SpreeExport(c *gin.Context) {

	var form ExportForm
	partNumber := c.Param("part")

	if c.Bind(&form) == nil {
		if len(partNumber) < 1 {
			c.JSON(400, gin.H{"message": "Missing part number.", "status": "error"})
			return
		}

		component, err := this.Components.Get(bson.M{"part_number": partNumber})

		if err != nil {
			c.JSON(404, gin.H{"message": "Invalid request, component not found.", "status": "error"})
			return
		}

		spree, err := component.Spree()

		if err != nil || spree == nil {
			c.JSON(500, gin.H{"message": err, "status": "error"})
			return
		}

		spree.UpdateStock(form.Stock)
		spree.UpdatePrice(form.Price)

		c.JSON(200, gin.H{"status": "okay"})
		return
	}

	c.JSON(400, gin.H{"message": "Could not bind payload to form.", "status": "error"})
}

type ExportForm struct {
	Stock bool    `json:"stock"`
	Price float64 `json:"price" binding:"required"`
}
