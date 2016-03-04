package components

import (
	"github.com/gin-gonic/gin"
)

func (this API) Search(c *gin.Context) {

	query := c.Query("q")
	ls, aggregation := this.Components.SearchComponents(query)

	c.JSON(200, gin.H{"facets": aggregation, "results": ls})
}
