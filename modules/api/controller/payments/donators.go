package payments

import (
	"github.com/gin-gonic/gin"
)

func (this API) GetTopDonators(c *gin.Context) {

	donators := this.User.GetTopDonators()

	c.JSON(200, donators)
}
