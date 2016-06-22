package builds

import (
//"github.com/gin-gonic/gin"
//"gopkg.in/jmcvetta/neoism.v1"

//"strconv"
)

/*func (a API) CreateAction(c *gin.Context) {

	var m BuildPayload

	if c.Bind(&m) == nil {

		db := a.Neoism
		params := neoism.Props{}
		parts := ``

		for _, partKind := range m.Parts {
			for index, option := range partKind.Options {

				pid := strconv.Itoa(index)
				slug := "slug" + pid
				price := "price" + pid
				label := "label" + pid

				parts += "MERGE (b)-[r:HAS_COMPONENT {label: {" + label + "}, price: {" + price + "}}]->(c:Component {slug: {" + slug + "}}) "
			}
		}

		query := neoism.CypherQuery{
			Statement: `
				CREATE (b:Build {name: {name}, price: {price}})
			`,
			// Use parameters instead of constructing a query string
			Parameters: neoism.Props{"name": q, "ctype": te},
			Result:     &data,
		}
	}
}*/

type PartPayload struct {
	Slug     string        `json:"slug" binding:"required"`
	Label    string        `json:"label" binding:"required"`
	Price    int           `json:"price" binding:"required"`
	Upgrades []PartPayload `json:"upgrades,omitempty"`
}

type PartMapPayload struct {
	Type    string        `json:"type" binding:"required"`
	Options []PartPayload `json:"options" binding:"required"`
}

/*type BuildPayload struct {
	Id    int              `json:"id"`
	Name  string           `json:"name" binding:"required"`
	Price int              `json:"price" binding:"required"`
	Parts []PartMapPayload `json:"parts" binding:"required"`
}*/
