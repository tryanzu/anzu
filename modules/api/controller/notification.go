package controller

import (
	"github.com/gin-gonic/gin"
	notify "github.com/tryanzu/core/board/notifications"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"strconv"
)

// Return user notifications
func Notifications(c *gin.Context) {
	var (
		take int = 10
		skip int = 0
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
