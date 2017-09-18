package post

import (
	notify "github.com/fernandez14/spartangeek-blacker/board/notifications"
	"github.com/fernandez14/spartangeek-blacker/core/events"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"gopkg.in/mgo.v2/bson"
)

// Bind event handlers for posts related actions...
func init() {
	events.On <- events.EventHandler{
		On: events.POSTS_NEW,
		Handler: func(e events.Event) error {
			post, err := FindId(deps.Container, e.Params["id"].(bson.ObjectId))
			if err != nil {
				return err
			}

			notify.Transmit <- notify.Socket{"feed", "action", map[string]interface{}{
				"fire":     "new-post",
				"category": post.Category.Hex(),
				"user_id":  post.UserId.Hex(),
				"id":       post.Id.Hex(),
				"slug":     post.Slug,
			}}

			// Ignore hard-coded category.
			if post.Category.Hex() == "55dc16593f6ba1005d000007" {
				return nil
			}

			return gaming.UserHasPublished(deps.Container, post.UserId)
		},
	}
}
