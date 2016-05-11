package comments

import (
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"github.com/fernandez14/spartangeek-blacker/modules/notifications"
	"github.com/fernandez14/spartangeek-blacker/modules/transmit"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"

	"regexp"
)

var legalSlug = regexp.MustCompile(`^([a-zA-Z0-9\-\.|/]+)$`)

type API struct {
	Feed          *feed.FeedModule                   `inject:""`
	Acl           *acl.Module                        `inject:""`
	Components    *components.Module                 `inject:""`
	Gaming        *gaming.Module                     `inject:""`
	Transmit      *transmit.Sender                   `inject:""`
	Mongo         *mongo.Service                     `inject:""`
	Errors        *exceptions.ExceptionsModule       `inject:""`
	Notifications *notifications.NotificationsModule `inject:""`
	Config        *config.Config                     `inject:""`
	S3            *s3.Bucket                         `inject:""`
}

type CommentForm struct {
	Content string `json:"content" binding:"required"`
}
