package checkout

import (
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

const ITEM_NOT_FOUND = "not-found"
const ITEM_NO_SELLER = "invalid-seller"
const ITEM_NOT_AVAILABLE = "cant-sell"
const ITEM_CHEAPER = "cheaper-now"
const ITEM_MORE_EXPENSIVE = "more-expensive"

type API struct {
	Store      *store.Module      `inject:""`
	Components *components.Module `inject:""`
	GCommerce  *gcommerce.Module  `inject:""`
	Mail       *mail.Module       `inject:""`
	User       *user.Module       `inject:""`
}

func (this API) getCartObject(c *gin.Context) *cart.Cart {

	obj, err := cart.Boot(cart.GinGonicSession{sessions.Default(c)})

	if err != nil {
		panic(err)
	}

	return obj
}

type CheckoutForm struct {
	Gateway string                 `json:"gateway" binding:"required"`
	ShipTo  bson.ObjectId          `json:"ship_to" binding:"required"`
	Order   bson.ObjectId          `json:"order_id"`
	Total   float64                `json:"total" binding:"required"`
	Meta    map[string]interface{} `json:"meta"`
}

type MassdropForm struct {
	Gateway   string                 `json:"gateway" binding:"required"`
	ProductId bson.ObjectId          `json:"product_id" binding:"required"`
	Quantity  int                    `json:"quantity" binding:"required"`
	Meta      map[string]interface{} `json:"meta"`
}

type CheckoutError struct {
	Type    string                 `json:"type"`
	Related bson.ObjectId          `json:"related_id"`
	Meta    map[string]interface{} `json:"data,omitempty"`
}
