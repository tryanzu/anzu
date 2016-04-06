package cart

import (
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/contrib/sessions"
)

type API struct {
	GCommerce  *gcommerce.Module  `inject:""`
	Components *components.Module `inject:""`
}

func (this API) getCart(c *gin.Context) *cart.Cart {

	obj, err := cart.Boot(cart.GinGonicSession{sessions.Default(c)})

	if err != nil {
		panic(err)
	}

	return obj
}

type CartAddForm struct {	
	Id string `json:"id" binding:"required"`
}