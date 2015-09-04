package handle

import (
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"github.com/gin-gonic/gin"
)

type GamingAPI struct {
	Gaming *gaming.Module `inject:""`
}

func (di *GamingAPI) GetRules(c *gin.Context) {

	rules := di.Gaming.Rules

	// Just return the previously loaded rules
	c.JSON(200, rules)
}
