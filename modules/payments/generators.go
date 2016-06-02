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

func (c *Create) SetIntent(name PaymentType) {
	c.Type = name.String()
}

func (c *Create) SetRelated(name string, id bson.ObjectId) {
	c.Related = name
	c.RelatedId = id
}

func (c *Create) SetProducts(ls []Product) {
	c.Products = ls

	for _, p := range ls {
		c.Total += float64(p.GetQuantity()) * p.GetPrice()
	}
}

func (c *Create) Purchase() (p *Payment, response map[string]interface{}, err error) {

	p = &Payment{
		Id:      bson.NewObjectId(),
		Type:    c.Type,
		Amount:  c.Total,
		UserId:  c.UserId,
		Gateway: c.G.GetName(),
		Status:  PAYMENT_CREATED,
		Created: time.Now(),
		Updated: time.Now(),
	}

	if c.Related != "" && c.RelatedId.Valid() {
		p.Related = c.Related
		p.RelatedId = c.RelatedId
	}

	response, err = c.G.Purchase(p, c)

	// Save payment information
	c.M.Mongo.Save(p)

	return
}
