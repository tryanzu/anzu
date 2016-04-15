package posts

import (
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Get(c *gin.Context) {

	var kind string
	var post *feed.Post
	var err error

	id := c.Params.ByName("id")

	if bson.IsObjectIdHex(id) {
		kind = "id"
	}

	if legalSlug.MatchString(id) && kind == "" {
		kind = "slug"
	}

	if kind == "" {
		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	if kind == "id" {
		post, err = this.Feed.Post(bson.ObjectIdHex(id))
	} else {
		post, err = this.Feed.Post(bson.M{"slug": id})
	}

	if err != nil {
		c.JSON(404, gin.H{"message": "Couldnt found post with that slug.", "status": "error"})
		return
	}

	// Needed data loading to show post
	post.LoadComments(-10, 0)
	post.LoadUsers()

	_, signed_in := c.Get("token")

	if signed_in {

		user_str_id := c.MustGet("user_id")
		user_id := bson.ObjectIdHex(user_str_id.(string))

		post.LoadVotes(user_id)

		go func(post *feed.Post, user_id string, signed_in bool) {

			defer this.Errors.Recover()

			if signed_in {

				by := bson.ObjectIdHex(user_id)

				post.Viewed(by)
			}

			post.UpdateRate()

			// Trigger gamification events (if needed)
			this.Gaming.Post(post).Review()

		}(post, user_str_id.(string), signed_in)
	}

	data := post.Data()

	true_count := this.Feed.TrueCommentCount(data.Id)
	data.Comments.Total = true_count

	c.JSON(200, data)
}
