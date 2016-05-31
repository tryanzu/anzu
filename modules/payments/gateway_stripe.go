package payments

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	"github.com/stripe/stripe-go/client"
)

type Stripe struct {
	Client      *client.API
	TokenSource string
	Description string
}

func (s *Stripe) GetName() string {
	return "stripe"
}

func (s *Stripe) SetOptions(o map[string]interface{}) {

	if source, exists := o["source"]; exists {
		s.TokenSource = source.(string)
	}

	if desc, exists := o["description"]; exists {
		s.Description = desc.(string)
	}
}

func (s *Stripe) Purchase(pay *Payment, c *Create) (map[string]interface{}, error) {

	cents := uint64(c.Total * 100.0)
	chargeParams := &stripe.ChargeParams{
		Amount:   cents,
		Currency: "mxn",
		Desc:     s.Description,
	}

	chargeParams.SetSource(s.TokenSource)

	if s.Client == nil {
		panic("Invalid stripe client pointer.")
	}

	stripeCharge, err := charge.New(chargeParams)
	response := map[string]interface{}{}

	// Keep some data about the stripe transaction inside payment struct
	pay.Meta = stripeCharge
	pay.GatewayId = stripeCharge.ID

	if err != nil {
		pay.Status = PAYMENT_ERROR
	} else {
		pay.Status = PAYMENT_SUCCESS
		response["stripe_id"] = stripeCharge.ID
	}

	return response, err
}

func (p *Stripe) CompletePurchase(pay *Payment, data map[string]interface{}) (map[string]interface{}, error) {

	response := map[string]interface{}{}

	return response, nil
}
