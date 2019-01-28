package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

func (this API) ResendConfirmation(c *gin.Context) {
	id := c.MustGet("userID").(bson.ObjectId)

	// Get the user using its id
	usr, err := user.FindId(deps.Container, id)
	if err != nil {
		panic(err)
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
