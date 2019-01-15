package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/comments"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

// Comments paginated fetch.
func Comments(c *gin.Context) {
	var (
		pid    = c.Param("post_id")
		limit  = 10
		offset = 0
		sort   = c.Query("sort")
		before *bson.ObjectId
		after  *bson.ObjectId
	)

	if n, err := strconv.Atoi(c.Query("limit")); err == nil && n <= 50 {
		limit = n
	}

	if n, err := strconv.Atoi(c.Query("offset")); err == nil {
		offset = n
	}

	if bid := c.Query("before"); len(bid) > 0 && bson.IsObjectIdHex(bid) {
		id := bson.ObjectIdHex(bid)
		before = &id
	}

	if bid := c.Query("after"); len(bid) > 0 && bson.IsObjectIdHex(bid) {
		id := bson.ObjectIdHex(bid)
		after = &id
	}

	list, err := comments.FetchBy(
		deps.Container,
		comments.Post(bson.ObjectIdHex(pid), limit, offset, sort == "reverse", before, after),
	)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	list, err = list.WithReplies(deps.Container, 5)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	list, err = list.WithUsers(deps.Container)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	empty := make(map[string]interface{}, 0)
	tables := map[string]interface{}{
		"votes":    empty,
		"comments": list.StrMap(),
	}

	if userID, exists := c.Get("userID"); exists {
		votes, err := list.VotesOf(deps.Container, userID.(bson.ObjectId))
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		tables["votes"] = votes.ValuesMap()
	}

	c.JSON(200, gin.H{"status": "okay", "list": list.IDList(), "hashtables": tables})
}

// NewComment pushes a new reply.
func NewComment(c *gin.Context) {
	var (
		kind = c.DefaultQuery("type", "post")
		cid  = bson.ObjectIdHex(c.Param("id"))
		form struct {
			Content string `json:"content" binding:"required"`
		}
	)

	if cid.Valid() == false {
		c.AbortWithError(500, errors.New("Invalid id for reply"))
		return
	}

	if err := c.BindJSON(&form); err != nil {
		c.AbortWithError(500, errors.New("Invalid kind of reply"))
		return
	}

	if kind != "post" && kind != "comment" {
		c.AbortWithError(500, errors.New("Invalid kind of reply"))
		return
	}

	usr := c.MustGet("user").(user.User)
	comment, err := comments.UpsertComment(deps.Container, comments.Comment{
		UserId:    usr.Id,
		Content:   form.Content,
		ReplyType: kind,
		ReplyTo:   cid,
	})

	if err != nil {
		c.AbortWithError(500, errors.New("Invalid kind of reply"))
		return
	}

	events.In <- events.PostComment(comment.Id)
	c.JSON(200, comment)
}

// UpdateComment pushes a new reply.
func UpdateComment(c *gin.Context) {
	var (
		cid  = bson.ObjectIdHex(c.Param("id"))
		form struct {
			Content string `json:"content" binding:"required"`
		}
	)

	if cid.Valid() == false {
		c.AbortWithError(500, errors.New("Invalid id for reply"))
		return
	}

	if err := c.BindJSON(&form); err != nil {
		c.AbortWithError(500, errors.New("Invalid reply body."))
		return
	}

	comment, err := comments.FindId(deps.Container, cid)
	if err != nil {
		c.AbortWithError(404, errors.New("Unknown comment to update."))
		return
	}

	comment.Content = form.Content
	updated, err := comments.UpsertComment(deps.Container, comment)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// Notify other processes...
	events.In <- events.UpdateComment(signs(c), comment.PostId, comment.Id)
	c.JSON(200, updated)
}

func DeleteComment(c *gin.Context) {
	var (
		cid = bson.ObjectIdHex(c.Param("id"))
	)

	if cid.Valid() == false {
		c.AbortWithError(500, errors.New("Invalid id for reply"))
		return
	}

	comment, err := comments.FindId(deps.Container, cid)
	if err != nil {
		c.AbortWithError(404, errors.New("unknown comment to delete"))
		return
	}

	err = comments.Delete(deps.Container, comment)
	if err != nil {
		panic(err)
		c.AbortWithError(500, err)
		return
	}
	/*user_str := c.MustGet("user_id")
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

	comment.Delete()*/

	// Notify events pool.
	events.In <- events.DeleteComment(signs(c), comment.PostId, comment.Id)
	c.JSON(200, gin.H{"status": "okay"})
}

func signs(c *gin.Context) events.UserSign {
	usr := c.MustGet("user").(user.User)
	sign := events.UserSign{
		UserID: usr.Id,
	}
	if r := c.Query("reason"); len(r) > 0 {
		sign.Reason = r
	}
	return sign
}
