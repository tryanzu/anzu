package payments

import (
	"github.com/dustin/go-humanize"
	"github.com/leebenson/paypal"
)

type Paypal struct {
	Client       *paypal.Client
	ReturnUrl    string
	CancelUrl    string
	Currency     string
	GDescription string // Global description for paypal transaction
	SDescription string // Soft description (see paypal docs)
}

func (p *Paypal) GetName() string {
	return "paypal"
}

func (p *Paypal) SetOptions(o map[string]interface{}) {

	if client, exists := o["client"]; exists {
		p.Client = client.(*paypal.Client)
	}

	if rurl, exists := o["return_url"]; exists {
		p.ReturnUrl = rurl.(string)
	}

	if curl, exists := o["cancel_url"]; exists {
		p.CancelUrl = curl.(string)
	}

	if d, exists := o["description"]; exists {
		p.GDescription = d.(string)
	}

	if sd, exists := o["soft_description"]; exists {
		p.SDescription = sd.(string)
	}

	if c, exists := o["currency"]; exists {
		p.Currency = c.(string)
	}
}

func (p *Paypal) generateTransactions(total float64, products []Product) []paypal.Transaction {

	var currency string
	var gdesc string
	var sdesc string

	first := products[0]

	if len(p.Currency) > 0 {
		currency = p.Currency
	} else {
		currency = first.GetCurrency()
	}

	if len(p.GDescription) > 0 {
		gdesc = p.GDescription
	} else {
		gdesc = first.GetDescription()
	}

	if len(p.SDescription) > 0 {
		sdesc = p.SDescription
	} else {
		sdesc = first.GetName()
	}

	items := []paypal.Item{}

	for _, p := range products {

		items = append(items, paypal.Item{
			Quantity:    p.GetQuantity(),
			Name:        p.GetName(),
			Currency:    p.GetCurrency(),
			Description: p.GetDescription(),
			Price:       humanize.FormatFloat("###.##", p.GetPrice()),
		})
	}

	t := []paypal.Transaction{
		{
			Amount: &paypal.Amount{
				Currency: currency,
				Total:    humanize.FormatFloat("###.##", total),
				/*Details: &paypal.Details{
					Shipping: "119.00",
					Subtotal: "116.00",
					Tax:      "3.00",
				},*/
			},
			Description: gdesc,
			ItemList: &paypal.ItemList{
				Items: items,
			},
			SoftDescriptor: sdesc,
		},
	}

	return t
}

func (p *Paypal) Purchase(pay *Payment, c *Create) (map[string]interface{}, error) {

	t := p.generateTransactions(c.Total, c.Products)

	payment := paypal.Payment{
		Intent: "sale",
		Payer: &paypal.Payer{
			PaymentMethod: "paypal",
		},
		Transactions: t,
		RedirectURLs: &paypal.RedirectURLs{
			CancelURL: p.CancelUrl,
			ReturnURL: p.ReturnUrl,
		},
	}

	client := p.Client

	if client == nil {
		panic("Invalid paypal client pointer.")
	}

	response := map[string]interface{}{}
	dopayment, err := client.CreatePayment(payment)

	// Keep some data about the paypal transaction inside payment struct
	pay.Meta = dopayment

	if err == nil {
		pay.GatewayId = dopayment.ID
	}

	if err != nil {
		pay.Status = PAYMENT_ERROR
	} else {
		pay.Status = PAYMENT_AWAITING

		for _, l := range dopayment.Links {
			if l.Rel == "approval_url" {
				response["approval_url"] = l.Href
			}
		}
	}

	return response, err
}

func (p *Paypal) CompletePurchase(pay *Payment, data map[string]interface{}) (map[string]interface{}, error) {

	if p.Client == nil {
		panic("Can't complete purchase if paypal client is not setup.")
	}

	payerId, exists := data["payer_id"]

	if !exists {
		panic("Can't complete purchase without payer_id in options.")
	}

	do, err := p.Client.ExecutePayment(pay.GatewayId, payerId.(string), nil)
	response := map[string]interface{}{}

	if err == nil {

		err := pay.UpdateStatus(PAYMENT_SUCCESS)

		if err != nil {
			panic(err)
		}

		for _, l := range do.Links {
			if l.Rel == "self" {
				response["self"] = l.Href
			}
		}
	} else {

		err := pay.UpdateStatus(PAYMENT_ERROR)

		if err != nil {
			panic(err)
		}
	}

	return response, err
}
