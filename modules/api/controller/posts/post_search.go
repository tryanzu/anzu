package posts

import (
	"github.com/gin-gonic/gin"
)

func (this API) Search(c *gin.Context) {

	query := c.Query("q")
	posts := this.Feed.SearchPosts(query)

	c.JSON(200, posts)
}
