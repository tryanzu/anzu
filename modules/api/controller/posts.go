package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/comments"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

// UpdatePost pushes a new reply.
func UpdatePost(c *gin.Context) {
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
