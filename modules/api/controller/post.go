package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

type PostAPI struct {
	Feed *feed.FeedModule `inject:""`
	Acl  *acl.Module      `inject:""`
}

func (self PostAPI) MarkCommentAsAnswer(c *gin.Context) {
	
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
	user := self.Acl.User(user_id)

	post, err := self.Feed.Post(id)

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

	c.JSON(200, gin.H{"status": "okay"})
}