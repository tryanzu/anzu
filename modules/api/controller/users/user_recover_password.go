package users

import (
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) RequestPasswordRecovery(c *gin.Context) {

	email := c.Query("email")

	if helpers.IsEmail(email) == false {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request, need valid email."})
		return
	} 

	// Get the user using its id
	usr, err := this.User.Get(bson.M{"email": email})

	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	go usr.SendRecoveryEmail()

	c.JSON(200, gin.H{"status": "okay"})
}
