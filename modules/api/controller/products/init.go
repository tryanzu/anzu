package products

import (
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
)

type API struct {
	Components *components.Module `inject:""`
	Feed       *feed.FeedModule   `inject:""`
	User       *user.Module       `inject:""`
	GCommerce  *gcommerce.Module  `inject:""`
}

type MassdropForm struct {
	Reference string `json:"reference"`
}
