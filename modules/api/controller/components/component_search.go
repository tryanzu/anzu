package components

import (
	"github.com/gin-gonic/gin"

	"strconv"
	"time"
)

func (this API) Search(c *gin.Context) {

	var limit int = 10
	var offset int = 0

	query := c.Query("q")
	kind := c.Query("category")
	in_store := c.Query("in_store") == "true"
	limitQuery := c.Query("limit")
	offsetQuery := c.Query("offset")

	if lq, err := strconv.Atoi(limitQuery); err == nil {
		limit = lq
	}

	if oq, err := strconv.Atoi(offsetQuery); err == nil {
		offset = oq
	}

	start := time.Now()
	ls, aggregation, count := this.Components.List(limit, offset, query, kind, in_store)
	elapsed := time.Since(start)

	c.JSON(200, gin.H{"limit": limit, "offset": offset, "facets": aggregation, "results": ls, "total": count, "elapsed": elapsed / time.Millisecond})
}
