package payments

import (
	"github.com/dustin/go-humanize"
	"github.com/leebenson/paypal"
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
		p.SDescription = s.(string)
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
			Price:       p.GetPrice(),
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
			Description: m.Description,
			ItemList: &paypal.ItemList{
				Items: []paypal.Item{
					{
						Quantity:    1,
						Name:        m.Description,
						Price:       humanize.FormatFloat("###.##", m.Amount),
						Currency:    "MXN",
						Description: m.Description,
						/*Tax:         "16.00",*/
					},
				},
			},
			SoftDescriptor: "SPARTANGEEK.COM",
		},
	}
}

func (p *Paypal) Charge(c Create) error {

	items := p.generateTransactions(c.Products)
	payment := paypal.Payment{
		Intent: "sale",
		Payer: &paypal.Payer{
			PaymentMethod: "paypal",
		},
		Transactions: []paypal.Transaction{
			{
				Amount: &paypal.Amount{
					Currency: "MXN",
					Total:    humanize.FormatFloat("###.##", m.Amount),
					/*Details: &paypal.Details{
						Shipping: "119.00",
						Subtotal: "116.00",
						Tax:      "3.00",
					},*/
				},
				Description: m.Description,
				ItemList: &paypal.ItemList{
					Items: []paypal.Item{
						{
							Quantity:    1,
							Name:        m.Description,
							Price:       humanize.FormatFloat("###.##", m.Amount),
							Currency:    "MXN",
							Description: m.Description,
							/*Tax:         "16.00",*/
						},
					},
				},
				SoftDescriptor: "SPARTANGEEK.COM",
			},
		},
		RedirectURLs: &paypal.RedirectURLs{
			CancelURL: baseUrl + "/donacion/error/",
			ReturnURL: baseUrl + "/donacion/exitosa/",
		},
	}

}
