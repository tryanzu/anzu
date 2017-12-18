package notifications

import (
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/modules/exceptions"
	"github.com/tryanzu/core/modules/user"
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
