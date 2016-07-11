package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

type OwnersAPI struct {
	User       *user.Module                 `inject:""`
	Components *components.Module           `inject:""`
	Errors     *exceptions.ExceptionsModule `inject:""`
}

// Push a owners entity relation
func (self OwnersAPI) Post(c *gin.Context) {

	var form OwnersPostForm

	kindParam := c.Param("kind")
	idParam := c.Param("id")
	usrParam := c.MustGet("user_id")
	usrId := bson.ObjectIdHex(usrParam.(string))

	if bson.IsObjectIdHex(idParam) && c.Bind(&form) == nil {

		if IsOwnStatusValid(kindParam, form.Status) {

			usr, err := self.User.Get(usrId)

			if err != nil {
				c.JSON(400, gin.H{"status": "error", "message": "Auth could not be performed, check token."})
				return
			}

			id := bson.ObjectIdHex(idParam)
			component, err := self.Components.Get(id)

			if err != nil {
				c.JSON(400, gin.H{"status": "error", "message": "Component could not be found, check id."})
				return
			}

			usr.Owns(form.Status, kindParam, component.Id)

			c.JSON(200, gin.H{"status": "okay"})
			return
		}
	}

	c.JSON(400, gin.H{"status": "error", "message": "Invalid vote request, check docs."})
	return
}

func (self OwnersAPI) Delete(c *gin.Context) {

	kindParam := c.Param("kind")
	idParam := c.Param("id")
	usrParam := c.MustGet("user_id")
	usrId := bson.ObjectIdHex(usrParam.(string))

	if bson.IsObjectIdHex(idParam) {

		usr, err := self.User.Get(usrId)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "Auth could not be performed, check token."})
			return
		}

		id := bson.ObjectIdHex(idParam)
		component, err := self.Components.Get(id)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "Component could not be found, check id."})
			return
		}

		usr.ROwns(kindParam, component.Id)

		c.JSON(200, gin.H{"status": "okay"})
		return
	}

	c.JSON(400, gin.H{"status": "error", "message": "Invalid delete vote request, check docs."})
	return
}

func IsOwnStatusValid(name string, status *string) bool {

	if name != "component" && name != "component-buy" {
		return false
	}

	if name == "component" && *status != "want-it" && *status != "have-it" && *status != "had-it" && status != nil {
		return false
	}

	if name == "component-buy" && *status != "yes" && *status != "maybe" && *status != "no" && *status != "wow" {
		return false
	}

	return true
}

type OwnersPostForm struct {
	Status *string `json:"status" binding:"required"`
}
