package interfaces

import (
	"github.com/fernandez14/spartangeek-blacker/model"
)

type NotificationBroadcaster interface {
	Send(message *model.UserFirebaseNotification) // TODO - Decouple the model from firebase implementation
}
