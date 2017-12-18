package comments

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/core/events"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Update(c *gin.Context) {
	var form CommentForm

	idstr := c.Params.ByName("id")
	if bson.IsObjectIdHex(idstr) == false {
		c.JSON(400, gin.H{"error": "Invalid request, no valid params.", "status": 701})
		return
	}

	if c.Bind(&form) == nil {

		id := bson.ObjectIdHex(idstr)
		user_str := c.MustGet("user_id")
		user_id := bson.ObjectIdHex(user_str.(string))
		usr := this.Acl.User(user_id)
		comment, err := this.Feed.GetComment(id)

		if err != nil {
			c.JSON(404, gin.H{"status": "error", "message": "Comment not found."})
			return
		}

		post := comment.GetPost()

		if usr.CanUpdateComment(comment, post) == false {
			c.JSON(400, gin.H{"message": "Can't update comment. Insufficient permissions.", "status": "error"})
			return
		}

		comment.Update(form.Content)

		events.In <- events.UpdateComment(post.Id, comment.Id)

		c.JSON(200, gin.H{"status": "okay", "message": comment.Content})
		return
	}
}
