package comments

import (
	"github.com/fernandez14/spartangeek-blacker/core/events"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Add(c *gin.Context) {
	var form CommentForm

	id := c.Params.ByName("id")
	if bson.IsObjectIdHex(id) == false {
		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	post, err := this.Feed.Post(bson.ObjectIdHex(id))
	if err != nil {
		c.JSON(404, gin.H{"status": "error", "message": "Post not found."})
		return
	}

	user_str := c.MustGet("user_id")
	user_id := bson.ObjectIdHex(user_str.(string))

	if c.Bind(&form) == nil {
		if post.IsLocked() {
			c.JSON(403, gin.H{"status": "error", "message": "Comments not longer allowed in this post."})
			return
		}

		user := this.Acl.User(user_id)

		if user.HasValidated() == false {
			c.JSON(403, gin.H{"status": "error", "message": "Not enough permissions."})
			return
		}

		comment := post.PushComment(form.Content, user_id)

		// Notify events pool.
		events.In <- events.PostComment(comment.Id)

		c.JSON(200, gin.H{"status": "okay", "message": comment.Content, "comment": comment, "position": comment.Position})
		return
	}

	c.JSON(401, gin.H{"error": "Not authorized.", "status": 704})
}
