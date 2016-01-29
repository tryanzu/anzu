package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

type ComponentAPI struct {
	Components *components.Module `inject:""`
	Feed       *feed.FeedModule   `inject:""`
	User       *user.Module       `inject:""`
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

	user_ref, signed_in := c.Get("user_id")

	if signed_in {

		user_id := bson.ObjectIdHex(user_ref.(string))
		usr, err := this.User.Get(user_id)

		if err == nil {

			// Track user viewing the component
			usr.TrackView("component", component.Id)
		}
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

	posts, err := this.Feed.LightPosts(bson.M{"related_components": component.Id, "deleted": bson.M{"$exists": false}})

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

// Delete Component's price
func (this ComponentAPI) DeletePrice(c *gin.Context) {

	slug := c.Param("slug")
	component, err := this.Components.Get(bson.M{"slug": slug})

	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid request, component not found.", "status": "error"})
		return
	}

	component.DeletePrice()

	c.JSON(200, gin.H{"status": "okay"})
	
}

type ComponentPriceUpdateForm struct {
	Price map[string]map[string]interface{} `json:"prices" binding:"required"`
}
