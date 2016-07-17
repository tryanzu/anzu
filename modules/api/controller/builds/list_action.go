package builds

import (
	"github.com/gin-gonic/gin"
)

func (a API) ListAction(c *gin.Context) {

	// Get pointer to builds service
	all := a.Builds.FindAll()

	c.JSON(200, gin.H{"builds": all})
}
