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

	count := 0

	if fcount := firebase.Child(target_path+"/count", nil, nil).Value(); fcount != nil {

		switch fcount.(type) {
		case int:
			count = fcount.(int)
		case float64:
			count = int(fcount.(float64))
		}

	}

	// Increase the notifications count
	firebase.Set(target_path+"/count", count+1, nil)

	// Send the notification to firebase straight forward
	firebase.Child(target_path+"/list", nil, nil).Push(message, nil)

	return
}
