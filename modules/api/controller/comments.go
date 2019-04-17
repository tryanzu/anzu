package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/comments"
	posts "github.com/tryanzu/core/board/posts"
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

	set, err := comments.FetchBy(
		deps.Container,
		comments.Post(bson.ObjectIdHex(pid), limit, offset, sort == "reverse", before, after),
	)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	list := set.List
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

	tables := map[string]interface{}{
		"votes":    map[string]interface{}{},
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

	c.JSON(200, gin.H{"status": "okay", "count": set.Count, "list": list.IDList(), "hashtables": tables})
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
		c.AbortWithError(500, errors.New("invalid reply body"))
		return
	}
	comment, err := comments.FindId(deps.Container, cid)
	if err != nil {
		c.AbortWithError(404, errors.New("unknown comment to update"))
		return
	}
	post, err := posts.FindId(deps.Container, comment.ReplyTo)
	if err != nil {
		c.AbortWithError(404, errors.New("unknown comment's post to update"))
		return
	}
	if perms(c).CanUpdateComment(comment.UserId, post.Category) == false {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Not allowed to perform this operation"})
		return
	}
	comment.Content = form.Content
	updated, err := comments.UpsertComment(deps.Container, comment)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// Notify other processes...
	events.In <- events.UpdateComment(signs(c), comment.ReplyTo, comment.Id)
	c.JSON(200, updated)
}

// DeleteComment endpoint handler.
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

	post, err := posts.FindId(deps.Container, comment.ReplyTo)
	if err != nil {
		c.AbortWithError(404, errors.New("unknown comment's post to update"))
		return
	}

	if perms(c).CanDeleteComment(comment.UserId, post.Category) == false {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Not allowed to perform this operation"})
		return
	}

	err = comments.Delete(deps.Container, comment)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	// Notify events pool.
	events.In <- events.DeleteComment(signs(c), comment.PostId, comment.Id)
	c.JSON(200, gin.H{"status": "okay"})
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
