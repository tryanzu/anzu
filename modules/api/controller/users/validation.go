package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (this API) ResendConfirmation(c *gin.Context) {
	id := c.MustGet("userID").(primitive.ObjectID)

	// Get the user using its id
	usr, err := user.FindId(deps.Container, id)
	if err != nil {
		if err == user.UserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to retrieve user"})
		return
	}

	if usr.Validated {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "message": "User has been validated already."})
		return
	}

	err = usr.ConfirmationEmail(deps.Container)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "okay"})
}
