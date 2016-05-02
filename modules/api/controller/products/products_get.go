package products

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Get(c *gin.Context) {

	slug := c.Param("id")

	if len(slug) < 1 {
		c.JSON(400, gin.H{"message": "Invalid request, need component slug.", "status": "error"})
		return
	}

	products := this.GCommerce.Products()
	product, err := products.GetByBson(bson.M{"slug": slug})

	if err != nil {
		c.JSON(404, gin.H{"message": "Invalid request, product not found.", "status": "error"})
		return
	}

	var user_id bson.ObjectId

	if _, signed_in := c.Get("token"); signed_in {

		user_str := c.MustGet("user_id")
		user_id = bson.ObjectIdHex(user_str.(string))
	}

	if user_id.Valid() {
		product.ShareRequesterUserId(user_id)
	}

	// Load Massdrop information (if exists)
	product.InitializeMassdrop()

	if user_id.Valid() {
		product.UserMassdrop(user_id)
	}

	c.JSON(200, product)
}
