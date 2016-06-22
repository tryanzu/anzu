package builds

import (
	"github.com/fernandez14/spartangeek-blacker/modules/builds"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (a API) GetAction(c *gin.Context) {

	var sessionId string
	var err error
	var userId *bson.ObjectId
	var build *builds.Build

	// Get pointer to builds service
	service := a.Builds

	if session, exists := c.Get("session_id"); exists {
		sessionId = session.(string)
	} else {
		c.JSON(400, gin.H{"status": "error", "message": "Could not validate request session."})
		return
	}

	if user, exists := c.Get("user_id"); exists {
		id := bson.ObjectIdHex(user.(string))
		userId = &id
	}

	// The get request could get an id for an specific build or none for the in-session build
	if c.Param("id") != "" {
		buildId := c.Param("id")
		build, err = service.FindByRef(buildId)

		if err != nil {
			c.JSON(404, gin.H{"status": "error", "message": "Could not get build by ref."})
		}
	} else {
		session := sessions.Default(c)
		bi := session.Get("build_ref")

		if bi == nil {
			// Attempt to find user/session build or create one in case theres none
			build = service.FindOrCreate(sessionId, userId)
		} else {
			// Attempt to get build using in-session ref
			build, err = service.FindByRef(bi.(string))

			// If last find fails (shall never) then fallback to user/session lookup
			if err != nil {
				build = service.FindOrCreate(sessionId, userId)
			}
		}

		if build != nil {
			session.Set("build_ref", build.Ref)
			session.Save()
		}
	}

	if build == nil {
		c.JSON(500, gin.H{"status": "error", "message": "Could not get healthy build reference."})
		return
	}

	c.JSON(200, build)
}
