package components

import (
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"gopkg.in/jmcvetta/neoism.v1"
)

type API struct {
	Components *components.Module `inject:""`
	Feed       *feed.FeedModule   `inject:""`
	User       *user.Module       `inject:""`
	Neoism     *neoism.Database   `inject:""`
}
