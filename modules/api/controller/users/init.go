package users

import (
	"github.com/tryanzu/core/modules/feed"
	"github.com/tryanzu/core/modules/user"
)

type API struct {
	Feed *feed.FeedModule `inject:""`
	User *user.Module     `inject:""`
}
