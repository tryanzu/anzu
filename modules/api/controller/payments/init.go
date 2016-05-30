package payments

import (
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/payments"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/olebedev/config"
)

type API struct {
	Gaming     *gaming.Module     `inject:""`
	Config     *config.Config     `inject:""`
	Store      *store.Module      `inject:""`
	Components *components.Module `inject:""`
	GCommerce  *gcommerce.Module  `inject:""`
	Payments   *payments.Module   `inject:""`
	Mail       *mail.Module       `inject:""`
	Mongo      *mongo.Service     `inject:""`
	User       *user.Module       `inject:""`
}

func (this API) GetPaypalGateway(options map[string]interface{}) *payments.Paypal {

	baseUrl, err := this.Config.String("application.siteUrl")

	if err != nil {
		panic("Could not get siteUrl from config.")
	}

	gateway := this.Payments.GetGateway("paypal")
	o := map[string]interface{}{
		"return_url": baseUrl + "/donacion/exitosa/",
		"cancel_url": baseUrl + "/donacion/error/",
		"currency":   "MXN",
	}

	// Merge options
	for key, value := range options {
		o[key] = value
	}

	gateway.SetOptions(o)

	return gateway.(*payments.Paypal)
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
