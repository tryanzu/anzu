package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/modules/gaming"
)

type GamingAPI struct {
	Gaming *gaming.Module `inject:""`
}

// Get gamification rules (roles, badges)
func (di *GamingAPI) GetRules(c *gin.Context) {

	rules := di.Gaming.GetRules()

	// Just return the previously loaded rules
	c.JSON(200, rules)
}
