package comments

import (
	"github.com/fernandez14/spartangeek-blacker/core/events"
	notify "github.com/fernandez14/spartangeek-blacker/core/notifications"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"gopkg.in/mgo.v2/bson"
)

// Bind event handlers for comment related actions...
func init() {
	events.On <- events.EventHandler{
		On: events.POSTS_COMMENT,
		Handler: func(e events.Event) error {
			comment, err := FindId(deps.Container, e.Params["id"].(bson.ObjectId))
			if err != nil {
				return err
			}

			notify.Database <- notify.Notification{
				UserId:    comment.UserId,
				Type:      "comment",
				RelatedId: comment.Id,
			}

			notify.Transmit <- notify.Socket{"feed", "action", map[string]interface{}{
				"fire":    "new-comment",
				"id":      comment.PostId.Hex(),
				"user_id": comment.UserId.Hex(),
			}}

			deps.Container.Transmit().Emit("feed", "action", params)

			// Ignore hard-coded category.
			if post.Category.Hex() == "55dc16593f6ba1005d000007" {
				return nil
			}

			return gaming.UserHasCommented(deps.Container, comment.UserId)
		},
	}
}
