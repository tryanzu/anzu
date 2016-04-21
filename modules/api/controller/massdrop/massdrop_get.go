package massdrop

import (
	"github.com/gin-gonic/gin"

	"strconv"
)

func (this API) Get(c *gin.Context) {

	var limit int = 10
	var offset int = 0

	limitq := c.Query("limit")
	offsetq := c.Query("offset")

	if lq, err := strconv.Atoi(limitq); err == nil {
		limit = lq
	}

	if oq, err := strconv.Atoi(offsetq); err == nil {
		offset = oq
	}

	products := this.GCommerce.Products()
	ls := products.GetMassdrops(limit, offset)

	c.JSON(200, gin.H{"limit": limit, "offset": offset, "results": ls})
}
