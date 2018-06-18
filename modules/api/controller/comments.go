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

	list, err := comments.FetchBy(deps.Container, comments.Post(bson.ObjectIdHex(pid), limit, offset, sort == "reverse", before))
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	list, err = list.WithReplies(deps.Container, 3)
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
		"comments": list.Map(),
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
