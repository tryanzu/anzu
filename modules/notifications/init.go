package notifications

import (
	"github.com/fernandez14/spartangeek-blacker/board/legacy/model"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"gopkg.in/mgo.v2/bson"
)

type NotificationsModule struct {
	Errors *exceptions.ExceptionsModule `inject:""`
	User   *user.Module                 `inject:""`
}

type MentionParseObject struct {
	Type          string
	RelatedNested string
	Content       string
	Title         string
	Author        bson.ObjectId
	Post          model.Post
}
