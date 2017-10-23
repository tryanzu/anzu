package users

import (
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
)

type API struct {
	Feed *feed.FeedModule `inject:""`
	User *user.Module     `inject:""`
}
