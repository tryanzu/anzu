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

	start := time.Now()
	ls, aggregation, count := this.Components.List(limit, offset, query, kind, in_store)
	elapsed := time.Since(start)

	c.JSON(200, gin.H{"limit": limit, "offset": offset, "facets": aggregation, "results": ls, "total": count, "elapsed": elapsed/time.Millisecond})
}
