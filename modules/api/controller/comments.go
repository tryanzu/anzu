package controller

import (
	"github.com/fernandez14/spartangeek-blacker/board/comments"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/gin-gonic/gin"
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

	c.JSON(200, list)
}
