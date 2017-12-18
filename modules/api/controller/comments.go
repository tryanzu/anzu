package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/comments"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"

	"strconv"
)

func Comments(c *gin.Context) {
	var (
		pid    string = c.Param("post_id")
		limit  int    = 10
		offset int    = 0
	)

	if n, err := strconv.Atoi(c.Query("limit")); err == nil && n <= 50 {
		limit = n
	}

	if n, err := strconv.Atoi(c.Query("offset")); err == nil && n > 0 {
		offset = n
	}

	list, err := comments.FetchBy(deps.Container, comments.Post(bson.ObjectIdHex(pid), limit, offset))
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

	c.JSON(200, list)
}
