package cart

import (
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/gin-gonic/gin"
)

func (this API) Get(c *gin.Context) {

	var items []cart.CartItem

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
