package events

import (
	"github.com/tryanzu/core/board/comments"
	notify "github.com/tryanzu/core/board/notifications"
	"github.com/tryanzu/core/board/posts"
	pool "github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/gaming"
	"gopkg.in/mgo.v2/bson"
)

// Bind event handlers for comment related actions...
func commentsEvents() {
	pool.On <- pool.EventHandler{
		On: pool.POSTS_COMMENT,
		Handler: func(e pool.Event) error {
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

	pool.On <- pool.EventHandler{
		On: pool.COMMENT_DELETE,
		Handler: func(e pool.Event) error {
			cid := e.Params["id"].(bson.ObjectId)
			pid := e.Params["post_id"].(bson.ObjectId)

			notify.Transmit <- notify.Socket{"feed", "action", map[string]interface{}{
				"fire": "delete-comment",
				"id":   pid.Hex(),
			}}

			notify.Transmit <- notify.Socket{"post", pid.Hex(), map[string]interface{}{
				"fire": "delete-comment",
				"id":   cid.Hex(),
			}}

			return nil
		},
	}

	pool.On <- pool.EventHandler{
		On: pool.COMMENT_UPDATE,
		Handler: func(e pool.Event) error {
			cid := e.Params["id"].(bson.ObjectId)
			pid := e.Params["post_id"].(bson.ObjectId)

			notify.Transmit <- notify.Socket{"post", pid.Hex(), map[string]interface{}{
				"fire": "comment-updated",
				"id":   cid.Hex(),
			}}

			return nil
		},
	}
}
