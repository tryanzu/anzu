package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

type OwnersAPI struct {
	User       *user.Module       `inject:""`
	Components *components.Module `inject:""`
	Errors     *exceptions.ExceptionsModule `inject:""`
}

// Push a owners entity relation 
func (self OwnersAPI) Post(c *gin.Context) {

	var form OwnersPostForm

	kindParam := c.Param("kind")
	idParam  := c.Param("id")
	usrParam := c.MustGet("user_id")
	usrId := bson.ObjectIdHex(usrParam.(string))

	if kindParam == "component" && bson.IsObjectIdHex(idParam) && c.Bind(&form) == nil {
		
		if form.Status == "want-it" || form.Status == "have-it" || form.Status == "had-it" {

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

			usr.Owns(form.Status, "component", component.Id)

			c.JSON(200, gin.H{"status": "okay"})
		}  
	}

	c.JSON(400, gin.H{"status": "error", "message": "Invalid vote request, check docs."})
	return
}

type OwnersPostForm struct {
	Status string `json:"status" binding:"required"`
}