package users

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/modules/helpers"
	"gopkg.in/mgo.v2/bson"
)

var PATCHABLE_FIELDS = []string{"onesignal_id"}

type PatchForm struct {
	Value string `form:"value" json:"value" binding:"required"`
}

func (this API) Patch(c *gin.Context) {

	var form PatchForm

	field := c.Param("field")
	id := c.MustGet("user_id")
	userId := bson.ObjectIdHex(id.(string))

	if exists, _ := helpers.InArray(field, PATCHABLE_FIELDS); exists && c.Bind(&form) == nil {
		usr, err := this.User.Get(userId)
		if err != nil {
			panic(err)
		}

		err = usr.Update(map[string]interface{}{field: form.Value})
		if err != nil {
			panic(err)
		}

		c.JSON(200, gin.H{"status": "okay"})
		return
	}

	c.JSON(400, gin.H{"status": "error", "message": "Invalid request."})
}
