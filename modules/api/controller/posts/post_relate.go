package posts

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Relate(c *gin.Context) {

	post_id := c.Param("id")
	related_id := c.Param("related_id")

	if !bson.IsObjectIdHex(post_id) || !bson.IsObjectIdHex(related_id) {

		c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
		return
	}

	id := bson.ObjectIdHex(post_id)
	component_id := bson.ObjectIdHex(related_id)

	// Validate post and related component
	post, err := this.Feed.Post(id)

	if err != nil {
		c.JSON(404, gin.H{"status": "error", "message": "Post not found."})
		return
	}

	component, err := this.Components.Get(component_id)

	if err != nil {

		c.JSON(404, gin.H{"status": "error", "message": "Component not found."})
		return
	}

	post.Attach(component)

	c.JSON(200, gin.H{"status": "okay"})
}
