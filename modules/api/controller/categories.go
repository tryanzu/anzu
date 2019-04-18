package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/categories"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/acl"
	"gopkg.in/mgo.v2/bson"
)

// Categories paginated fetch.
func Categories(c *gin.Context) {
	tree := categories.MakeTree(deps.Container)
	if sid, exists := c.Get("userID"); exists {
		uid := sid.(bson.ObjectId)
		auth := acl.LoadedACL.User(uid)
		tree = tree.CheckWrite(auth.CanWrite)
	}
	c.JSON(200, tree)
}
