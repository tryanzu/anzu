package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	notify "github.com/tryanzu/core/board/notifications"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
)

// Notifications from current authenticated user.
func Notifications(c *gin.Context) {
	var (
		take = 10
		skip = 0
	)

	if n, err := strconv.Atoi(c.Query("take")); err == nil && n <= 50 {
		take = n
	}

	if n, err := strconv.Atoi(c.Query("skip")); err == nil && n > 0 {
		skip = n
	}

	usr := c.MustGet("user").(user.User)
	list, err := notify.FetchBy(deps.Container, notify.UserID(usr.Id, take, skip))
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	batch, err := list.Humanize(deps.Container)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	if len(batch) == 0 {
		c.JSON(200, make([]string, 0))
		return
	}

	err = user.ResetNotifications(deps.Container, usr.Id)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, batch)
}
