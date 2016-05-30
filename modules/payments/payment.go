package payments

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"time"
)

const PAYMENT_ERROR = "error"
const PAYMENT_SUCCESS = "confirmed"
const PAYMENT_AWAITING = "awaiting"
const PAYMENT_CREATED = "created"

type Payment struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	UserId    bson.ObjectId `bson:"user_id" json:"user_id"`
	Type      string        `bson:"type" json:"type"`
	Amount    float64       `bson:"amount" json:"amount"`
	Gateway   string        `bson:"gateway" json:"gateway"`
	GatewayId string        `bson:"gateway_id" json:"gateway_id"`
	Meta      interface{}   `bson:"gateway_response,omitempty" json:"gateway_response,omitempty"` // TODO - Move it to another collection
	Status    string        `bson:"status" json:"status"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
	Updated   time.Time     `bson:"updated_at" json:"updated_at"`

	di *Module
}

func (p *Payment) SetDI(di *Module) {
	p.di = di
}

func (p *Payment) CompletePurchase(d map[string]interface{}) (map[string]interface{}, error) {

	if p.di == nil {
		panic("No DI injected.")
	}

	g := p.di.GetGateway(p.Gateway)
	res, err := g.CompletePurchase(p, d)

	return res, err
}

func (p *Payment) UpdateStatus(status string) error {

	if p.di == nil {
		panic("No DI injected.")
	}

	database := p.di.Mongo.Database
	err := database.C("payments").Update(bson.M{"_id": p.Id}, bson.M{"$set": bson.M{"updated_at": time.Now(), "status": status}})

	if err == nil {
		p.Status = status
	}

	return err
}

func (p *Payment) Save(db *mgo.Database) error {

	err := db.C("payments").Insert(p)

	return err
}
