package notifications

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"gopkg.in/mgo.v2/bson"

	"fmt"
	"strconv"
	"time"
)

func (self *NotificationsModule) Comment(post_slug, post_title string, position int, post_id, post_user, user_id bson.ObjectId) {

	defer self.Errors.Recover()

	usr, err := self.User.Get(user_id)

	if err != nil {
		panic(fmt.Sprintf("Could not get user while notifying comment (post_id: %v, user_id: %v, position: %v)", post_id, user_id, position))
	}

	// Construct the notification message
	title := fmt.Sprintf("Nuevo comentario de **%s**", usr.Name())
	message := post_title

	notification := model.UserFirebaseNotification{
		UserId:       post_user,
		RelatedId:    post_id,
		RelatedExtra: post_slug,
		Position:     strconv.Itoa(position),
		Title:        title,
		Text:         message,
		Related:      "comment",
		Seen:         false,
		Image:        usr.Data().Image,
		Created:      time.Now(),
		Updated:      time.Now(),
	}

	broadcaster := self.Broadcaster
	broadcaster.Send(notification)
}
