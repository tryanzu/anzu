package payments

import (
	"github.com/dustin/go-humanize"
	"github.com/leebenson/paypal"

	"fmt"
)

type Paypal struct {
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

func (p *Paypal) Charge(c Create) error {

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

	fmt.Printf("%v\n", payment)

	return nil
}
