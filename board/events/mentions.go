package events

import (
	notify "github.com/tryanzu/core/board/notifications"
	ev "github.com/tryanzu/core/core/events"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Bind event handlers for posts related actions...
func mentionEvents() {
	ev.On <- ev.EventHandler{
		On: ev.NEW_MENTION,
		Handler: func(e ev.Event) error {
			userID := e.Params["user_id"].(primitive.ObjectID)
			relatedID := e.Params["related_id"].(primitive.ObjectID)
			users := e.Params["users"].([]primitive.ObjectID)
			related := e.Params["related"].(string)

			// Create notification
			if related == "comment" {
				notify.Database <- notify.Notification{
					UserId:    userID,
					Type:      "mention",
					RelatedId: relatedID,
					Users:     users,
				}
			}
			if related == "chat" {
				notify.Database <- notify.Notification{
					UserId:    userID,
					Type:      "chat",
					RelatedId: relatedID,
					Users:     users,
				}
			}
			return nil
		},
	}
}
