package http

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

// Load site config into middlewares pipe context.
func SiteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		config := deps.Container.Config()
		siteName, err := config.String("site.name")
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		description, err := config.String("site.description")
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		url, err := config.String("site.url")
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.Set("siteName", siteName)
		c.Set("siteDescription", description)
		c.Set("siteUrl", url)
		c.Next()
	}
}

// Limit number of simultaneous connections.
func MaxAllowed(n int) gin.HandlerFunc {
	sem := make(chan bool, n)
	acquire := func() { sem <- true }
	release := func() { <-sem }
	return func(c *gin.Context) {
		acquire()       // before request
		defer release() // after request
		c.Next()
	}
}

func TitleMiddleware(title string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("siteName", title)
		c.Next()
	}
}

// User middleware loads signed user data for further use.
func UserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.MustGet("user_id").(string)
		oid := bson.ObjectIdHex(sid)

		// Attempt to retrieve user data otherwise abort request.
		usr, err := user.FindId(deps.Container, oid)
		if err != nil {
			c.AbortWithError(412, err)
			return
		}
		sign := events.UserSign{
			UserID: oid,
		}
		if r := c.Query("reason"); len(r) > 0 {
			sign.Reason = r
		}

		c.Set("sign", sign)
		c.Set("user", usr)
		c.Next()
	}
}
