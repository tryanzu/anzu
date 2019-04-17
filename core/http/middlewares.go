package http

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/acl"
	"gopkg.in/mgo.v2/bson"
)

// Load site config into middlewares pipe context.
func SiteMiddleware() gin.HandlerFunc {
	legacy := deps.Container.Config()
	siteName, err := legacy.String("site.name")
	if err != nil {
		log.Panicf("site.name not found in config")
	}

	description, err := legacy.String("site.description")
	if err != nil {
		log.Panicf("site.description not found in config")
	}

	url, err := legacy.String("site.url")
	if err != nil {
		log.Panicf("site.url not found in config")
	}

	return func(c *gin.Context) {
		c.Set("config", config.C.Copy())
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

func Can(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		users := acl.LoadedACL.User(c.MustGet("userID").(bson.ObjectId))
		if users.Can(permission) == false {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Not allowed to perform this operation"})
			return
		}
		c.Set("acl", users)
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
