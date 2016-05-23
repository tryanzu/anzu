package payments

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"time"
)

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
}

func (p *Payment) Save(db *mgo.Database) error {

	err := db.C("payments").Insert(p)

	return err
}
