package content

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/notifications"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/xuyu/goredis"
)

type Module struct {
	Mongo         *mongo.Service                     `inject:""`
	Errors        *exceptions.ExceptionsModule       `inject:""`
	S3            *s3.Bucket                         `inject:""`
	Config        *config.Config                     `inject:""`
	Notifications *notifications.NotificationsModule `inject:""`
	Redis         *goredis.Redis                     `inject:""`
}
