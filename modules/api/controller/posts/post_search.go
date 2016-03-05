package posts

import (
	"github.com/gin-gonic/gin"

	"time"
)

func (this API) Search(c *gin.Context) {

	query := c.Query("q")

	start := time.Now()
	posts, count := this.Feed.SearchPosts(query)
	elapsed := time.Since(start)

	c.JSON(200, gin.H{"results": posts, "total": count, "elapsed": elapsed/time.Millisecond})
}
