package gcommerce

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	"gopkg.in/mgo.v2/bson"

	"errors"
	"math"
	"time"
)

type GatewayStripe struct {
	di    *Module
	order *Order
	meta  map[string]interface{}
}

// Set DI instance
func (this *GatewayStripe) SetDI(di *Module) {
	this.di = di
}

func (this *GatewayStripe) SetOrder(order *Order) {
	this.order = order
}

func (this *GatewayStripe) SetMeta(meta map[string]interface{}) {
	this.meta = meta
}

func (this *GatewayStripe) Charge(amount float64) error {

	database := this.di.Mongo.Database

	// Setup stripe private key always
	stripe.Key = this.di.StripeKey

	// Check meta data integrity
	reference, exists := this.meta["reference"].(string)

	if !exists {
		return errors.New("no-reference")
	}

	token, exists := this.meta["token"].(string)

	if !exists {
		return errors.New("invalid-token")
	}

	cents := uint64(amount * 100)
	address := this.order.GetRelatedAddress()
	order_address := this.order.Shipping.Address

	chargeParams := &stripe.ChargeParams{
		Amount:   cents,
		Currency: "mxn",
		Desc:     "Pago del pedido #" + reference,
		Shipping: &stripe.ShippingDetails{
		    Name: address.Recipient,
		    Phone: address.Phone,
		    Address: stripe.Address{
		      Line1: order_address.Line1,
		      Line2: order_address.Line2,
		      City: order_address.City,
		      Country: order_address.Country,
		      State: order_address.State,
		      Zip: order_address.PostalCode,
		    },
		},
	}

	chargeParams.SetSource(token)
	chargeParams.AddMeta("order_ref", reference)

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

		status := Status{
			this.order.Status,
			make(map[string]interface{}),
			this.order.Updated,
		}

		err = database.C("gcommerce_orders").Update(bson.M{"_id": this.order.Id}, bson.M{"$set": bson.M{"status": ORDER_PAYMENT_ERROR}, "$push": bson.M{"statuses": status}})

		if err != nil {
			panic(err)
		}

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
		default:
			return errors.New("gateway-error")
		}
	}

	status := Status{
		this.order.Status,
		make(map[string]interface{}),
		this.order.Updated,
	}

	err = database.C("gcommerce_orders").Update(bson.M{"_id": this.order.Id}, bson.M{"$set": bson.M{"status": ORDER_CONFIRMED}, "$push": bson.M{"statuses": status}})

	if err != nil {
		panic(err)
	}

	err = database.C("gcommerce_transactions").Insert(transaction)

	if err != nil {
		panic(err)
	}

	return nil
}

func (this *GatewayStripe) ModifyPrice(p float64) float64 {
	return p * 1.042
}

func (this *GatewayStripe) AdjustPrice(p float64) float64 {
	return math.Ceil(p + 4)
}
