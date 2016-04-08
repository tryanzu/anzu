package products

import (
	"github.com/gin-gonic/gin"

	"strconv"
	"time"
)

func (this API) Search(c *gin.Context) {

	var limit int = 10
	var offset int = 0

	query := c.Query("q")
	kind  := c.Query("type")
	category := c.Query("category")
	limitq := c.Query("limit")
	offsetq := c.Query("offset")

	if lq, err := strconv.Atoi(limitq); err == nil {
		limit = lq
	}

	if oq, err := strconv.Atoi(offsetq); err == nil {
		offset = oq
	}

	products := this.GCommerce.Products()
	start := time.Now()
	ls, facets, count, err := products.GetList(limit, offset, query, category, kind)
	if err != nil {
		panic(err)
	}

	elapsed := time.Since(start)

	c.JSON(200, gin.H{"limit": limit, "offset": offset, "facets": facets, "results": ls, "total": count, "elapsed": elapsed/time.Millisecond})
}
