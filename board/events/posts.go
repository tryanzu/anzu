package events

import (
	notify "github.com/tryanzu/core/board/notifications"
	posts "github.com/tryanzu/core/board/posts"
	ev "github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

// Bind event handlers for posts related actions...
func postsEvents() {
	ev.On <- ev.EventHandler{
		On: ev.POSTS_NEW,
		Handler: func(e ev.Event) error {
			post, err := posts.FindId(deps.Container, e.Params["id"].(bson.ObjectId))
			if err != nil {
				return err
			}

			notify.Transmit <- notify.Socket{
				Chan:   "feed",
				Action: "action",
				Params: map[string]interface{}{
					"fire":     "new-post",
					"category": post.Category.Hex(),
					"user_id":  post.UserId.Hex(),
					"id":       post.Id.Hex(),
					"slug":     post.Slug,
				},
			}

			return nil
		},
	}

	ev.On <- ev.EventHandler{
		On: ev.POST_VIEW,
		Handler: func(e ev.Event) error {
			post, err := posts.FindId(deps.Container, e.Params["id"].(bson.ObjectId))
			if err != nil {
				return err
			}

			err = posts.TrackView(deps.Container, post.Id, e.Sign.UserID)
			return err
		},
	}

	ev.On <- ev.EventHandler{
		On: ev.POSTS_REACHED,
		Handler: func(e ev.Event) error {
			list := e.Params["list"].([]bson.ObjectId)
			err := posts.TrackReachedList(deps.Container, list, e.Sign.UserID)
			return err
		},
	}
}
