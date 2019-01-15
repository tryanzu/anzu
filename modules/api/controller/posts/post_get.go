package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/modules/feed"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Get(c *gin.Context) {
	var (
		kind string
		post *feed.Post
		err  error
	)

	id := c.Params.ByName("id")
	if bson.IsObjectIdHex(id) {
		kind = "id"
	}

	if legalSlug.MatchString(id) && kind == "" {
		kind = "slug"
	}

	if kind == "" {
		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	if kind == "id" {
		post, err = this.Feed.Post(bson.ObjectIdHex(id))
	} else {
		post, err = this.Feed.Post(bson.M{"slug": id})
	}

	if err != nil {
		c.JSON(404, gin.H{"message": "Couldnt found post.", "status": "error"})
		return
	}

	// Needed data loading to show post
	post.LoadComments(10, -10)
	post.LoadUsers()

	if sid, exists := c.Get("userID"); exists {
		uid := sid.(bson.ObjectId)
		post.LoadVotes(uid)

		// Notify about view.
		events.In <- events.PostView(signs(c), post.Id)
	}

	post.LoadUsersHashtables()
	data := post.Data()
	data.Comments.Total = this.Feed.TrueCommentCount(data.Id)

	c.JSON(200, data)
}

func signs(c *gin.Context) events.UserSign {
	usr := c.MustGet("userID").(bson.ObjectId)
	sign := events.UserSign{
		UserID: usr,
	}
	if r := c.Query("reason"); len(r) > 0 {
		sign.Reason = r
	}
	return sign
}
