package posts

import (
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/transmit"
	"github.com/fernandez14/spartangeek-blacker/mongo"
)

type API struct {
	Feed       *feed.FeedModule   `inject:""`
	Acl        *acl.Module        `inject:""`
	Components *components.Module `inject:""`
	Transmit   *transmit.Sender   `inject:""`
	Mongo      *mongo.Service     `inject:""`
}