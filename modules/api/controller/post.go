package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/transmit"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

	"strconv"
)

type PostAPI struct {
	Feed       *feed.FeedModule   `inject:""`
	Acl        *acl.Module        `inject:""`
	Components *components.Module `inject:""`
	Transmit   *transmit.Sender   `inject:""`
}

func (this PostAPI) MarkCommentAsAnswer(c *gin.Context) {

	post_id := c.Param("id")
	comment_pos := c.Param("comment")

	if bson.IsObjectIdHex(post_id) == false {

		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	id := bson.ObjectIdHex(post_id)
	comment_index, err := strconv.Atoi(comment_pos)

	if err != nil {

		c.JSON(400, gin.H{"message": "Invalid request, comment index not valid.", "status": "error"})
		return
	}

	user_str_id := c.MustGet("user_id")
	user_id := bson.ObjectIdHex(user_str_id.(string))
	user := this.Acl.User(user_id)

	post, err := this.Feed.Post(id)

	if err != nil {

		c.JSON(404, gin.H{"status": "error", "message": "Post not found."})
		return
	}

	if user.CanSolvePost(post.Data()) == false {

		c.JSON(400, gin.H{"message": "Can't update post. Insufficient permissions", "status": "error"})
		return
	}

	if post.Data().Solved == true {

		c.JSON(400, gin.H{"status": "error", "message": "Already solved."})
		return
	}

	comment, err := post.Comment(comment_index)

	if err != nil {

		c.JSON(400, gin.H{"message": "Invalid request, comment index not valid.", "status": "error"})
		return
	}

	comment.MarkAsAnswer()

	go func(carrier *transmit.Sender, id bson.ObjectId) {

		carrierParams := map[string]interface{}{
			"fire": "best-answer",
			"id": id.Hex(),
		} 

		carrier.Emit("feed", "action", carrierParams)

	}(this.Transmit, post.Data().Id)

	c.JSON(200, gin.H{"status": "okay"})
}

func (this PostAPI) Relate(c *gin.Context) {

	post_id := c.Param("id")
	related_id := c.Param("related_id")

	if !bson.IsObjectIdHex(post_id) || !bson.IsObjectIdHex(related_id) {

		c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
		return
	}

	id := bson.ObjectIdHex(post_id)
	component_id := bson.ObjectIdHex(related_id)

	// Validate post and related component
	post, err := this.Feed.Post(id)

	if err != nil {

		c.JSON(404, gin.H{"status": "error", "message": "Post not found."})
		return
	}

	component, err := this.Components.Get(component_id)

	if err != nil {

		c.JSON(404, gin.H{"status": "error", "message": "Component not found."})
		return
	}

	post.Attach(component)

	c.JSON(200, gin.H{"status": "okay"})
}
