package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

// Users paginated fetch.
func Users(c *gin.Context) {
	var (
		limit  = 10
		sort   = c.Query("sort")
		before *bson.ObjectId
		after  *bson.ObjectId
	)
	if n, err := strconv.Atoi(c.Query("limit")); err == nil && n <= 50 {
		limit = n
	}

	if bid := c.Query("before"); len(bid) > 0 && bson.IsObjectIdHex(bid) {
		id := bson.ObjectIdHex(bid)
		before = &id
	}

	if bid := c.Query("after"); len(bid) > 0 && bson.IsObjectIdHex(bid) {
		id := bson.ObjectIdHex(bid)
		after = &id
	}

	set, err := user.FetchBy(
		deps.Container,
		user.Page(limit, sort == "reverse", before, after),
	)
	if err != nil {
		panic(err)
	}
	c.JSON(200, set)
}
