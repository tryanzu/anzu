package payments

import (
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
)

type API struct {
	Store      *store.Module      `inject:""`
	Components *components.Module `inject:""`
	GCommerce  *gcommerce.Module  `inject:""`
	Mail       *mail.Module       `inject:""`
	Mongo      *mongo.Service     `inject:""`
	User       *user.Module       `inject:""`
}

type PlacePayload struct {
	Type        string  `json:"type" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	Description string  `json:"description" binding:"required"`
}

type ExecutePayload struct {
	PaymentId string `json:"paymentID" binding:"required"`
	PayerId   string `json:"payerID" binding:"required"`
}
