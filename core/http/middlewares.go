package http

import (
	"github.com/fernandez14/spartangeek-blacker/core/user"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

// User middleware loads signed user data for further use.
func UserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.MustGet("user_id").(string)
		oid := bson.ObjectIdHex(sid)

		// Attempt to retrieve user data otherwise abort request.
		usr, err := user.FindId(deps.Container, oid)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.Set("user", usr)
		c.Next()
	}
}
