package users

import (
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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

func (this API) UpdatePasswordFromToken(c *gin.Context) {

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

	if valid == false {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request, need valid token."})
		return
	}

	var form model.UserProfileForm

	if c.BindWith(&form, binding.JSON) == nil {

		if form.Password != "" {

			usr, err := this.User.GetUserFromRecoveryToken(token)

			if err != nil {
				c.JSON(400, gin.H{"status": "error", "message": err.Error()})
				return
			}

			usr.Update(map[string]interface{}{"password": form.Password})

			c.JSON(200, gin.H{"status": "okay"})
			return
		}
	}

	c.JSON(400, gin.H{"status": "error", "message": "Invalid request."})
}
