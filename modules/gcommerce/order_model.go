package gcommerce

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

const ORDER_AWAITING string = "awaiting"

type Order struct {
	Id        bson.ObjectId          `bson:"_id,omitempty" json:"id,omitempty"`
	
	// Information about the order status
	Status    string                 `bson:"status" json:"status"`
	Statuses  []Status               `bson:"statuses" json:"statuses"`

	// Information about the user/customer
	UserId    bson.ObjectId          `bson:"customer_id" json:"customer_id"`
	
	// Information about the shipping
	Shipping  Shipping               `bson:"shipping" json:"shipping"`

	// Information about the order itself
	Items     []Item                 `bson:"items" json:"items"` 
	Total     float64                `bson:"total" json:"total"` 
	Gateway   string                 `bson:"gateway" json:"gateway"` 
	Meta      map[string]interface{} `bson:"meta" json:"meta"`
	Created   time.Time              `bson:"created_at" json:"created_at"`
	Updated   time.Time              `bson:"updated_at" json:"updated_at"`

	di        *Module
}

type Item struct {
	Name        string  `bson:"name" json:"name"`
	Image       string  `bson:"image" json:"image"`
	Description string  `bson:"description" json:"description"`
	Price       float64 `bson:"price" json:"price"`
	Quantity    int `bson:"quantity" json:"quantity"`
	Meta        map[string]interface{} `bson:"meta" json:"meta"`
}

type Status struct {
	Name     string                 `bson:"name" json:"name"`
	Meta     map[string]interface{} `bson:"meta" json:"meta"`
	Created  time.Time              `bson:"created_at" json:"created_at"`
}

type Shipping struct {
	Type    string  `bson:"type" json:"type"`
	Price   float64 `bson:"price" json:"price"`
	Meta    map[string]interface{} `bson:"meta" json:"meta"`
	Address Address `bson:"address" json:"address"`
}

type Address struct {
	Country    string `bson:"country" json:"country"`
	City       string `bson:"city" json:"city"`
	Line1      string `bson:"line1" json:"line1"`
	Line2      string `bson:"line2" json:"line3"`
	PostalCode string `bson:"postal_code" json:"postal_code"`
	State      string `bson:"state" json:"state"`
	Extra	   string `bson:"extra" json:"extra"`
} 