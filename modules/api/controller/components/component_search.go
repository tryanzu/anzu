package components

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/jmcvetta/neoism.v1"

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

func (this API) Lookup(c *gin.Context) {

	q := c.Query("q")
	te := c.Query("type")
	neo := this.Neoism

	data := []struct {
		ID       int
		Name     string
		FullName string
		Brand    string
		Type     string
	}{}

	cType := ""

	if len(te) > 0 {
		cType = "AND n.type = {ctype} "
	}

	query := neoism.CypherQuery{
		Statement: `
			MATCH (n:Component)
			WHERE n.full_name CONTAINS {name}` + cType + `
			RETURN ID(n) AS ID, n.full_name as FullName, n.type as Type, n.name as Name, n.manufacturer as Brand
			LIMIT 15
		`,
		// Use parameters instead of constructing a query string
		Parameters: neoism.Props{"name": q, "ctype": te},
		Result:     &data,
	}

	err := neo.Cypher(&query)

	if err != nil {
		panic(err)
	}

	c.JSON(200, gin.H{"results": data})
}
