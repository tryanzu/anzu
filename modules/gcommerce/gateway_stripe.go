package gcommerce

import (
	"github.com/fernandez14/go-siftscience"
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
	//stripe.Key = this.di.StripeKey

	// Check meta data integrity
	reference, exists := this.meta["reference"].(string)

	if !exists {
		return errors.New("no-reference")
	}

	session_id, exists := this.meta["session_id"].(string)

	if !exists {
		return errors.New("no-session-id")
	}

	token, exists := this.meta["token"].(string)

	if !exists {
		return errors.New("invalid-token")
	}

	var address *CustomerAddress
	var order_address *Address

	_, addressless := this.meta["addressless"]

	customer := this.order.GetCustomer()
	usr := customer.GetUser()
	micros := int64((amount * 100) * 10000)
	cents := uint64(amount * 100)

	chargeParams := &stripe.ChargeParams{
		Amount:   cents,
		Currency: "mxn",
		Desc:     "Pago del pedido #" + reference,
	}

	if !addressless {

		address = this.order.GetRelatedAddress()
		order_address = this.order.Shipping.Address

		chargeParams.Shipping = &stripe.ShippingDetails{
			Name:  address.Recipient,
			Phone: address.Phone,
			Address: stripe.Address{
				Line1:   order_address.Line1,
				Line2:   order_address.Line2,
				City:    order_address.City,
				Country: order_address.Country,
				State:   order_address.State,
				Zip:     order_address.PostalCode,
			},
		}
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

	siftTransaction := map[string]interface{}{
		"$session_id":       session_id,
		"$user_id":          usr.Data().Id.Hex(),
		"$user_email":       usr.Email(),
		"$transaction_type": "$sale",
		"$amount":           micros,
		"$currency_code":    "MXN",
		"$order_id":         reference,
		"$transaction_id":   ch.ID,
		"$payment_method": map[string]interface{}{
			"$payment_type":    "$credit_card",
			"$payment_gateway": "$stripe",
			"$stripe_token":    token,
		},
	}

	if !addressless {

		// Siftscience transaction
		siftAddress := map[string]interface{}{
			"$name":      address.Recipient,
			"$phone":     address.Phone,
			"$address_1": address.Line1(),
			"$address_2": address.Line2(),
			"$city":      order_address.City,
			"$region":    order_address.State,
			"$country":   "MX",
			"$zipcode":   order_address.PostalCode,
		}

		siftTransaction["$billing_address"] = siftAddress
		siftTransaction["$shipping_address"] = siftAddress

	}

	if err != nil {

		stripeErr := err.(*stripe.Error)

		siftTransaction["$transaction_status"] = "$failure"
		err := gosift.Track("$transaction", siftTransaction)

		if err != nil {
			panic(err)
		}

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

	siftTransaction["$transaction_status"] = "$success"
	err = gosift.Track("$transaction", siftTransaction)

	if err != nil {
		panic(err)
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
