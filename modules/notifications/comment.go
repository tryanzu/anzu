package notifications

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"gopkg.in/mgo.v2/bson"

	"fmt"
	"time"
)

func (self *NotificationsModule) Comment(post *feed.Post, comment *feed.Comment, user_id bson.ObjectId) {

	defer self.Errors.Recover()

	usr, err := self.User.Get(user_id)

	if err != nil {
		panic(fmt.Sprintf("Could not get user while notifying comment (post_id: %v, user_id: %v, comment_id: %v)", post.Id, comment.Id, user_id))
	}

	// Construct the notification message
	title := fmt.Sprintf("Nuevo comentario de **%s**", usr.Name())
	message := post.Title

	notification := model.UserFirebaseNotification{
		UserId:       user_id,
		RelatedId:    post.Id,
		RelatedExtra: post.Slug,
		Position:     comment.Position,
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
