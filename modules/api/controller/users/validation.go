package users

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) ResendConfirmation(c *gin.Context) {

	user_str := c.MustGet("user_id")
	user_id := bson.ObjectIdHex(user_str.(string))

	// Get the user using its id
	usr, err := this.User.Get(user_id)

	if err != nil {
		panic(err)
	}

	if usr.IsValidated() {
		c.JSON(409, gin.H{"status": "error", "message": "User has been confirmed already."})
		return
	}

	err = usr.SendConfirmationEmail()

	if err != nil {
		c.JSON(429, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "okay"})
}
