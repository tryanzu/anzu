package chat

import (
	"github.com/fernandez14/spartangeek-blacker/core/user"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/gin-gonic/gin"

	"time"
)

func SendMessage(c *gin.Context) {
	var body struct {
		Channel string `json:"channel" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if c.BindJSON(&body) == nil {
		user := c.MustGet("user").(user.User)

		deps.Container.Transmit().Emit("chat", body.Channel, map[string]interface{}{
			"content":   body.Content,
			"user_id":   user.Id,
			"username":  user.UserName,
			"avatar":    user.Image,
			"timestamp": time.Now().Unix(),
		})

		c.AbortWithStatus(200)
		return
	}
}
