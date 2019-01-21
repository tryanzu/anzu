package events

import (
	notify "github.com/tryanzu/core/board/notifications"
	ev "github.com/tryanzu/core/core/events"
	"gopkg.in/mgo.v2/bson"
)

// Bind event handlers for posts related actions...
func mentionEvents() {
	ev.On <- ev.EventHandler{
		On: ev.NEW_MENTION,
		Handler: func(e ev.Event) error {
			userID := e.Params["user_id"].(bson.ObjectId)
			relatedID := e.Params["related_id"].(bson.ObjectId)
			users := e.Params["users"].([]bson.ObjectId)

			// Create notification
			notify.Database <- notify.Notification{
				UserId:    userID,
				Type:      "mention",
				RelatedId: relatedID,
				Users:     users,
			}
			return nil
		},
	}
}
