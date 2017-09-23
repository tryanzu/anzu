package events

import (
	"github.com/fernandez14/spartangeek-blacker/board/comments"
	notify "github.com/fernandez14/spartangeek-blacker/board/notifications"
	"github.com/fernandez14/spartangeek-blacker/board/posts"
	ev "github.com/fernandez14/spartangeek-blacker/core/events"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"gopkg.in/mgo.v2/bson"
)

// Bind event handlers for comment related actions...
func commentsEvents() {
	ev.On <- ev.EventHandler{
		On: ev.POSTS_COMMENT,
		Handler: func(e ev.Event) error {
			comment, err := comments.FindId(deps.Container, e.Params["id"].(bson.ObjectId))
			if err != nil {
				return err
			}

			post, err := post.FindId(deps.Container, comment.PostId)
			if err != nil {
				return err
			}

			if post.UserId != comment.UserId {
				notify.Database <- notify.Notification{
					UserId:    post.UserId,
					Type:      "comment",
					RelatedId: comment.Id,
					Users:     []bson.ObjectId{comment.UserId},
				}
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
