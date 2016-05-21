package payments

import (
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/leebenson/paypal"
	"github.com/olebedev/config"
)

type API struct {
	Gaming     *gaming.Module     `inject:""`
	Config     *config.Config     `inject:""`
	Store      *store.Module      `inject:""`
	Components *components.Module `inject:""`
	GCommerce  *gcommerce.Module  `inject:""`
	Mail       *mail.Module       `inject:""`
	Mongo      *mongo.Service     `inject:""`
	User       *user.Module       `inject:""`
}

func (this API) GetPaypalClient() *paypal.Client {

	clientID, err := this.Config.String("ecommerce.paypal.clientID")

	if err != nil {
		panic("Could not get config data to initialize paypal client. (cid)")
	}

	secret, err := this.Config.String("ecommerce.paypal.secret")

	if err != nil {
		panic("Could not get config data to initialize paypal client. (secret)")
	}

	sandbox, err := this.Config.Bool("ecommerce.paypal.sandbox")

	if err != nil {
		panic("Could not get config data to initialize paypal client. (sd)")
	}

	var r string = paypal.APIBaseSandBox

	if !sandbox {
		r = paypal.APIBaseLive
	}

	client := paypal.NewClient(clientID, secret, r)

	return client
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
