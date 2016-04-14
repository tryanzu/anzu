package products

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Massdrop(c *gin.Context) {

	id := c.Param("id")
	user_str := c.MustGet("user_id")
	user_id := bson.ObjectIdHex(user_str.(string))

	if !bson.IsObjectIdHex(id) {
		c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
		return
	}

	products := this.GCommerce.Products()
	product_id := bson.ObjectIdHex(id)
	product, err := products.GetById(product_id)

	if err != nil {
		c.JSON(404, gin.H{"message": "Invalid request, product not found.", "status": "error"})
		return
	}

	// Load Massdrop information (if exists)
	product.InitializeMassdrop()

	if product.Massdrop == nil {
		c.JSON(400, gin.H{"message": "Invalid request, no massdrop available.", "status": "error"})
		return
	}

	var form MassdropForm

	if c.Bind(&form) == nil {

		toggle, err := product.MassdropInterested(user_id, form.Reference)

		if err != nil {

			c.JSON(400, gin.H{"message": err.Error(), "status": "error"})
			return
		}

		c.JSON(200, gin.H{"status": "okay", "interested": toggle})
		return
	}

	c.JSON(400, gin.H{"message": "Invalid request, no massdrop available.", "status": "error"})
}
