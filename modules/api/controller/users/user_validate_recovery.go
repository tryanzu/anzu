package users

import (
	"github.com/gin-gonic/gin"
)

func (this API) ValidatePasswordRecovery(c *gin.Context) {

	token := c.Param("token")

	if len(token) < 4 {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request, need valid token."})
		return
	}

	valid, err := this.User.IsValidRecoveryToken(token)

	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "okay", "valid": valid})
}
