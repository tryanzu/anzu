package content

import (
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/tryanzu/core/modules/exceptions"
	"github.com/tryanzu/core/modules/notifications"
	"github.com/xuyu/goredis"
)

type Module struct {
	Errors        *exceptions.ExceptionsModule       `inject:""`
	S3            *s3.Bucket                         `inject:""`
	Config        *config.Config                     `inject:""`
	Notifications *notifications.NotificationsModule `inject:""`
	Redis         *goredis.Redis                     `inject:""`
}
