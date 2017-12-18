package comments

import (
	"github.com/fernandez14/spartangeek-blacker/core/events"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Delete(c *gin.Context) {

	idstr := c.Params.ByName("id")

	if bson.IsObjectIdHex(idstr) == false {
		c.JSON(400, gin.H{"error": "Invalid request, no valid params.", "status": 701})
		return
	}

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

	if usr.CanDeleteComment(comment, post) == false {
		c.JSON(400, gin.H{"message": "Can't delete comment. Insufficient permissions.", "status": "error"})
		return
	}

	comment.Delete()

	events.In <- events.DeleteComment(post.Id, comment.Id)
	/**
	go func(post_id bson.ObjectId, comment_id bson.ObjectId) {

		carrierParams := map[string]interface{}{
			"fire": "delete-comment",
			"id":   post_id.Hex(),
		}

		deps.Container.Transmit().Emit("feed", "action", carrierParams)

		carrierParams = map[string]interface{}{
			"fire": "delete-comment",
			"id":   comment_id.Hex(),
		}

		deps.Container.Transmit().Emit("post", post_id.Hex(), carrierParams)

	}(post.Id, comment.Id)*/

	c.JSON(200, gin.H{"status": "okay"})
	return
}
