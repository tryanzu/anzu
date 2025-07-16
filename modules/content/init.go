package content

import (
	"github.com/olebedev/config"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/exceptions"
	"github.com/tryanzu/core/modules/notifications"
	"github.com/xuyu/goredis"
)

type Module struct {
	Errors        *exceptions.ExceptionsModule       `inject:""`
	S3            *deps.S3Service                    `inject:""`
	Config        *config.Config                     `inject:""`
	Notifications *notifications.NotificationsModule `inject:""`
	Redis         *goredis.Redis                     `inject:""`
}
