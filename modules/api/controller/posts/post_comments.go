package posts

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

	"strconv"
)

func (this API) GetPostComments(c *gin.Context) {
	post_id := c.Param("id")
	offsetQuery := c.Query("offset")
	limitQuery := c.Query("limit")
	commentStr := c.Query("id")

	var offset int = 0
	var limit int = 0
	var cid bson.ObjectId

	if bson.IsObjectIdHex(post_id) == false {
		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	id := bson.ObjectIdHex(post_id)
	offsetC, err := strconv.Atoi(offsetQuery)

	if offsetQuery != "" && err != nil {
		c.JSON(400, gin.H{"message": "Invalid request, offset not valid.", "status": "error"})
		return
	}

	limitC, err := strconv.Atoi(limitQuery)

	if limitQuery != "" && err != nil {
		c.JSON(400, gin.H{"message": "Invalid request, limit not valid.", "status": "error"})
		return
	}

	if commentStr != "" && bson.IsObjectIdHex(commentStr) {
		cid = bson.ObjectIdHex(commentStr)
	}

	offset = offsetC
	limit = limitC

	post, err := this.Feed.Post(id)

	if err != nil {
		c.JSON(404, gin.H{"message": "Invalid request, post not found.", "status": "error"})
		return
	}

	// Needed data loading to show post
	if cid.Valid() {
		post.LoadCommentById(cid)
	} else {
		post.LoadComments(limit, offset)
	}

	post.LoadUsers()

	_, signed_in := c.Get("token")

	if signed_in {

		user_str_id := c.MustGet("user_id")
		user_id := bson.ObjectIdHex(user_str_id.(string))

		post.LoadVotes(user_id)
	}

	data := post.Data()

	true_count := this.Feed.TrueCommentCount(data.Id)
	data.Comments.Total = true_count

	c.JSON(200, data)
}
