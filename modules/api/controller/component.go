package controller 

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/gin-gonic/gin"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
)

type ComponentAPI struct {
	Components *components.Module `inject:""`
	Feed       *feed.FeedModule `inject:""`
}

// Get component by slug
func (this ComponentAPI) Get(c *gin.Context) {

	slug := c.Param("id")

	if len(slug) < 1 {
		c.JSON(400, gin.H{"message": "Invalid request, need component slug.", "status": "error"})
		return
	}

	component, err := this.Components.Get(bson.M{"slug": slug})

	if err != nil {
		c.JSON(404, gin.H{"message": "Invalid request, component not found.", "status": "error"})
		return
	}

	c.JSON(200, component.GetData())
}

// Get component's related posts
func (this ComponentAPI) GetPosts(c *gin.Context) {

	id := c.Param("id")

	if !bson.IsObjectIdHex(id) {

		c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
		return
	}

	component_id := bson.ObjectIdHex(id)
	component, err := this.Components.Get(component_id)

	if err != nil {
		c.JSON(404, gin.H{"message": "Invalid request, component not found.", "status": "error"})
		return
	}

	posts, err := this.Feed.LightPosts(bson.M{"related_components": component.Id})

	if err != nil {

		c.JSON(200, []string{})
		return
	}

	posts = this.Feed.FulfillBestAnswer(posts)

	c.JSON(200, posts)
}

// Update component's price
func (this ComponentAPI) UpdatePrice(c *gin.Context) {
	
	var form ComponentPriceUpdateForm

	slug := c.Param("slug")

	if c.BindJSON(&form) == nil {

		component, err := this.Components.Get(bson.M{"slug": slug})

		if err != nil {
			c.JSON(400, gin.H{"message": "Invalid request, component not found.", "status": "error"})
			return
		}

		component.UpdatePrice(form.Price)

		c.JSON(200, gin.H{"status": "okay"})
	}
}

type ComponentPriceUpdateForm struct {
	Price  map[string]float64 `json:"price" binding:"required"`
}