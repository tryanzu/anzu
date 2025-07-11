package users

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/modules/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var PATCHABLE_FIELDS = []string{"onesignal_id"}

type PatchForm struct {
	Value string `form:"value" json:"value" binding:"required"`
}

func (this API) Patch(c *gin.Context) {

	var form PatchForm

	field := c.Param("field")
	id := c.MustGet("user_id")
	userId, err := primitive.ObjectIDFromHex(id.(string))
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid user ID."})
		return
	}

	if exists, _ := helpers.InArray(field, PATCHABLE_FIELDS); exists && c.Bind(&form) == nil {
		usr, err := this.User.Get(userId)
		if err != nil {
			c.JSON(404, gin.H{"status": "error", "message": "User not found."})
			return
		}

		err = usr.Update(map[string]interface{}{field: form.Value})
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "Failed to update user."})
			return
		}

		c.JSON(200, gin.H{"status": "okay"})
		return
	}

	c.JSON(400, gin.H{"status": "error", "message": "Invalid request."})
}
