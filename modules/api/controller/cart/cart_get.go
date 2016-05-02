package cart

import (
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Get(c *gin.Context) {

	var items []cart.CartItem

	// Initialize cart library
	err := this.getCart(c).Bind(&items)

	if _, signed_in := c.Get("token"); signed_in {

		user_str := c.MustGet("user_id")
		user_id := bson.ObjectIdHex(user_str.(string))

		usr := this.GCommerce.GetCustomerFromUser(user_id)
		cart, err := usr.GetCart()

		if len(cart) > 0 && err == nil {
			for _, item := range cart {
				items = append(items, item.Item)
			}
		}
	}

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
