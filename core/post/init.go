package post

import (
	"github.com/fernandez14/spartangeek-blacker/core/events"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"gopkg.in/mgo.v2/bson"
)

// Bind event handlers for posts related actions...
func init() {
	events.On(events.POSTS_NEW, func(e events.Event) error {
		post, err := FindId(deps.Container, e.Params["id"].(bson.ObjectId))
		if err != nil {
			return err
		}

		params := map[string]interface{}{
			"fire":     "new-post",
			"category": post.Category.Hex(),
			"user_id":  post.UserId.Hex(),
			"id":       post.Id.Hex(),
			"slug":     post.Slug,
		}

		deps.Container.Transmit().Emit("feed", "action", params)

		// Ignore hard-coded category.
		if publish.Category.Hex() == "55dc16593f6ba1005d000007" {
			return nil
		}

		return gaming.UserHasPublished(deps.Container, post.UserId)
	})
}
