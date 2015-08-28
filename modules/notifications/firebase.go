package notifications

import (
	"github.com/cosn/firebase"
	"github.com/fernandez14/spartangeek-blacker/model"
)

type FirebaseBroadcaster struct {
	Firebase *firebase.Client `inject:""`
}

func (broadcaster FirebaseBroadcaster) Send(message *model.UserFirebaseNotification) {

	var target_notification model.UserFirebaseNotifications

	// Firebase path target
	firebase := broadcaster.Firebase
	target_path := "users/" + message.UserId.Hex() + "/notifications"

	// TODO - As the notifications increases this will slow down the whole process, change this
	target_ref := firebase.Child(target_path, nil, &target_notification)

	// Increase the notifications count
	target_ref.Set("count", target_notification.Count+1, nil)

	// Send the notification to firebase straight forward
	target_ref.Child("list", nil, nil).Push(message, nil)

	return
}
