package payments

import (
	"gopkg.in/mgo.v2/bson"

	"time"
)

const PAYMENT_DONATION string = "donation"
const PAYMENT_ORDER string = "order"

type Create struct {
	G         Gateway
	M         *Module
	Total     float64
	Related   string
	RelatedId bson.ObjectId
	UserId    bson.ObjectId
	Type      string
	Products  []Product
}

func (m *Module) Create(g Gateway) Create {
	return Create{G: g, M: m}
}

func (c *Create) SetUser(id bson.ObjectId) {
	c.UserId = id
}

func (c *Create) SetIntent(name string) {
	c.Type = name
}

func (c *Create) SetProducts(ls []Product) {
	c.Products = ls

	for _, p := range ls {
		c.Total += float64(p.GetQuantity()) * p.GetPrice()
	}
}

func (c *Create) Purchase() (p *Payment, response map[string]interface{}, err error) {

	p = &Payment{
		Type:    c.Type,
		Amount:  c.Total,
		UserId:  c.UserId,
		Gateway: c.G.GetName(),
		Status:  "created",
		Created: time.Now(),
		Updated: time.Now(),
	}

	response, err = c.G.Purchase(p, c)

	// Save payment information
	c.M.Mongo.Save(p)

	return
}
