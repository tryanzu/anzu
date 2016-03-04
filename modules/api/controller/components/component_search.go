package components

import (
	"github.com/gin-gonic/gin"

	"strconv"
)

func (this API) Search(c *gin.Context) {

	var limit int = 10
	var offset int = 0

	query := c.Query("q")
	kind  := c.Query("type")
	limitQuery := c.Query("limit")
	offsetQuery := c.Query("offset")

	if lq, err := strconv.Atoi(limitQuery); err == nil {
		limit = lq
	}

	if oq, err := strconv.Atoi(offsetQuery); err == nil {
		offset = oq
	}

	ls, aggregation := this.Components.List(limit, offset, query, kind)

	c.JSON(200, gin.H{"limit": limit, "offset": offset, "facets": aggregation, "results": ls})
}
