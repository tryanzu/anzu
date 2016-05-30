package payments

import (
	"gopkg.in/mgo.v2/bson"
)

const PAYMENT_DONATION string = "donation"
const PAYMENT_ORDER string = "order"

type Create struct {
	G         Gateway
	Total     float64
	Related   string
	RelatedId bson.ObjectId
	UserId    bson.ObjectId
	Type      string
	Products  []Product
}

func (m *Module) Create(g Gateway) Create {
	return Create{G: g}
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

func (c *Create) Purchase() {

}
