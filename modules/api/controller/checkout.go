package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/gin-gonic/gin"
)

type CheckoutAPI struct {
	Store *store.Module `inject:""`
}

func (this CheckoutAPI) Place(c *gin.Context) {

	//var form OrderForm
	stripe.Key = "sk_test_81pQu0m3my2V2ERPW0MMAOml"

	chargeParams := &stripe.ChargeParams{
	  Amount: 400,
	  Currency: "mxn",
	  Desc: "Charge for test@example.com",
	}

	chargeParams.SetSource("tok_17CXDqKinZpZZUA2KjAW5KIy")
	ch, err := charge.New(chargeParams)
}

type OrderForm struct {
	Gateway  string       `json:"gateway" binding:"required"`
	Delivery DeliveryForm `json:"delivery" binding:"required"`	
	Meta     map[string]interface{} `json:"meta"`
}

type DeliveryForm struct {
	State   string `json:"state" binding:"required"`
	City    string `json:"city" binding:"required"`
	Zipcode string `json:"zipcode" binding:"required"`
	AddressLine1 string `json:"address_line1" binding:"required"`
	AddressLine2 string `json:"address_line2" binding:"required"`
}