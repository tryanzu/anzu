package deals

import (
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"gopkg.in/mgo.v2/bson"
)

type API struct {
	Components *components.Module `inject:""`
	Feed       *feed.FeedModule   `inject:""`
	User       *user.Module       `inject:""`
	GCommerce  *gcommerce.Module  `inject:""`
	Store      *store.Module      `inject:""`
}

type InvoiceForm struct {
	Id    bson.ObjectId `json:"id" binding:"required"`
	Total float64       `json:"total" binding:"required"`
	RFC   string        `json:"rfc" binding:"required"`
	Name  string        `json:"fiscal_name" binding:"required"`
}
