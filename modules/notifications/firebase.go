package notifications

import (
	"github.com/cosn/firebase"
	"github.com/fernandez14/spartangeek-blacker/model"
)

type FirebaseBroadcaster struct {
	Firebase *firebase.Client `inject:""`
}

func (broadcaster FirebaseBroadcaster) Send(message model.UserFirebaseNotification) {

	// Firebase path target
	firebase := broadcaster.Firebase
	target_path := "users/" + message.UserId.Hex() + "/notifications"

	root := firebase.Child(target_path, nil, nil)
	count := 0

	if fcount := root.Child("count", nil, nil).Value(); fcount != nil {
		count = fcount.(int)
	}

	// Increase the notifications count
	root.Set("count", count+1, nil)

	// Send the notification to firebase straight forward
	root.Child("list", nil, nil).Push(message, nil)

	return
}
