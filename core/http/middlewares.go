package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/acl"
	"gopkg.in/mgo.v2/bson"
)

// SiteMiddleware loads site config into middlewares pipe context.
func SiteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bucket := sessions.Default(c)
		if token := bucket.Get("jwt"); token != nil {
			c.Set("jwt", token.(string))
			bucket.Delete("jwt")
			bucket.Save()
		}
		cnf := config.C.Copy()
		c.Set("config", cnf)
		c.Set("siteName", cnf.Site.Name)
		c.Set("siteDescription", cnf.Site.Description)
		c.Set("siteUrl", cnf.Site.Url)
		c.Next()
	}
}

// Limit number of simultaneous connections.
func MaxAllowed(n int) gin.HandlerFunc {
	sem := make(chan struct{}, n)
	acquire := func() { sem <- struct{}{} }
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
		if !users.Can(permission) {
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

		if usr.Banned && usr.BannedUntil != nil {
			if time.Now().Before(*usr.BannedUntil) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"status":  "error",
					"message": "You are banned for now... check again later!",
					"until":   usr.BannedUntil,
				})
				return
			}
		}

		c.Set("sign", sign)
		c.Set("user", usr)
		c.Next()
	}
}
