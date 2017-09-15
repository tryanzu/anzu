package comments

import (
	notify "github.com/fernandez14/spartangeek-blacker/board/notifications"
	"github.com/fernandez14/spartangeek-blacker/core/events"
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

			return gaming.UserHasCommented(deps.Container, comment.UserId)
		},
	}
}
