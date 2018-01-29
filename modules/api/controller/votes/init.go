package votes

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/tryanzu/core/board/votes"
	"github.com/tryanzu/core/modules/acl"
	"github.com/tryanzu/core/modules/exceptions"
	"github.com/tryanzu/core/modules/feed"
	"github.com/tryanzu/core/modules/gaming"
	"github.com/tryanzu/core/modules/notifications"
	"github.com/tryanzu/core/modules/user"
)

type API struct {
	Feed          *feed.FeedModule                   `inject:""`
	Acl           *acl.Module                        `inject:""`
	Gaming        *gaming.Module                     `inject:""`
	Errors        *exceptions.ExceptionsModule       `inject:""`
	Notifications *notifications.NotificationsModule `inject:""`
	User          *user.Module                       `inject:""`
	Config        *config.Config                     `inject:""`
	S3            *s3.Bucket                         `inject:""`
}

type CommentForm struct {
	Direction string `json:"direction" binding:"required"`
}

func (c CommentForm) VoteType() votes.VoteType {
	if c.Direction == "up" {
		return 1
	}

	return -1
}
