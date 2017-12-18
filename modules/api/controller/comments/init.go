package comments

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/tryanzu/core/modules/acl"
	"github.com/tryanzu/core/modules/exceptions"
	"github.com/tryanzu/core/modules/feed"
	"github.com/tryanzu/core/modules/gaming"
	"github.com/tryanzu/core/modules/notifications"

	"regexp"
)

var legalSlug = regexp.MustCompile(`^([a-zA-Z0-9\-\.|/]+)$`)

type API struct {
	Feed          *feed.FeedModule                   `inject:""`
	Acl           *acl.Module                        `inject:""`
	Gaming        *gaming.Module                     `inject:""`
	Errors        *exceptions.ExceptionsModule       `inject:""`
	Notifications *notifications.NotificationsModule `inject:""`
	Config        *config.Config                     `inject:""`
	S3            *s3.Bucket                         `inject:""`
}

type CommentForm struct {
	Content string `json:"content" binding:"required"`
}
