package gcommerce

import (
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Cart struct {
	Id         bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	CustomerId bson.ObjectId `bson:"customer_id" json:"customer_id"`
	Item       cart.CartItem `bson:"item" json:"item"`
	Created    time.Time     `bson:"created_at" json:"created_at"`
	Updated    time.Time     `bson:"updated_at" json:"updated_at"`
}
