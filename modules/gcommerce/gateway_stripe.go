package gcommerce

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	
	"errors"
	"time"
	"math"
)

type GatewayStripe struct {
	di    *Module
	order *Order
}

// Set DI instance
func (this *GatewayStripe) SetDI(di *Module) {
	this.di = di
}

func (this *GatewayStripe) SetOrder(order *Order) {
	this.order = order
}

func (this *GatewayStripe) Charge(amount float64) error {

	database := this.di.Mongo.Database

	// Setup stripe private key always
	stripe.Key = this.di.StripeKey

	cents := uint64(amount * 100)
	chargeParams := &stripe.ChargeParams{
		Amount:   cents,
		Currency: "mxn",
		Desc:     "Test description",
	}

	chargeParams.SetSource(&stripe.CardParams{
		Name:   "Go Stripe",
		Number: "4242424242424242",
		Month:  "10",
		Year:   "20",
	})
	ch, err := charge.New(chargeParams)

	transaction := &Transaction{
		OrderId:  this.order.Id,
		Gateway:  "stripe",
		Response: ch,
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	if err != nil {

		stripeErr := err.(*stripe.Error)

		transaction.Error = stripeErr
		database.C("gcommerce_transactions").Insert(transaction)

		switch stripeErr.Code {
		case stripe.IncorrectNum:
			return errors.New("gateway-incorrect-num")
		case stripe.InvalidNum:
			return errors.New("gateway-invalid-num")
		case stripe.InvalidExpM:
			return errors.New("gateway-invalid-exp-m")
		case stripe.InvalidExpY:
			return errors.New("gateway-invalid-exp-y")
		case stripe.InvalidCvc:
			return errors.New("gateway-invalid-cvc")
		case stripe.ExpiredCard:
			return errors.New("gateway-expired-card")
		case stripe.IncorrectCvc:
			return errors.New("gateway-incorrect-cvc")
		case stripe.IncorrectZip:
			return errors.New("gateway-incorrect-zip")
		case stripe.CardDeclined:
			return errors.New("gateway-card-declined")
		case stripe.Missing:
			return errors.New("gateway-stripe-missing")
		case stripe.ProcessingErr:
			return errors.New("gateway-processing-err")
		}
	}

	database.C("gcommerce_transactions").Insert(transaction)

	return nil
}

func (this *GatewayStripe) ModifyPrice(p float64) float64 {
	return p * 1.042
}

func (this *GatewayStripe) AdjustPrice(p float64) float64 {
	return math.Ceil(p + 3.5)
}
