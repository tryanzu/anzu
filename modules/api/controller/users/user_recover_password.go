package users

import (
	"github.com/fernandez14/spartangeek-blacker/core/user"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func (this API) RequestPasswordRecovery(c *gin.Context) {
	if helpers.IsEmail(c.Query("email")) == false {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request, need valid email."})
		return
	}

	usr, err := user.FindEmail(deps.Container, c.Query("email"))
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	err = usr.RecoveryPasswordEmail(deps.Container)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

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
