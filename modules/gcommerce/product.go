package gcommerce

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Product struct {
	Id          bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name"`
	Description string                 `bson:"description" json:"description"`
	Slug        string                 `bson:"slug" json:"slug"`
	Image       string                 `bson:"image" json:"image"`
	Images      []string               `bson:"images" json:"images"`
	Type        string                 `bson:"type" json:"type"`
	Category    string                 `bson:"category" json:"category"`
	CategoryId  bson.ObjectId          `bson:"category_id,omitempty" json:"category_id,omitempty"`
	Price       float64                `bson:"price" json:"price"`
	Shipping    float64                `bson:"shipping_cost" json:"shipping_cost"`
	Search      string 				   `bson:"search" json:"search"`
	Relevance   int 				   `bson:"relevance" json:"relevance"`
	Attrs       map[string]interface{} `bson:"attributes" json:"attributes"`
	Created     time.Time              `bson:"created_at" json:"created_at"`
	Updated     time.Time              `bson:"updated_at" json:"updated_at"`

	di *Module
	Massdrop    *Massdrop              `bson:"massdrop,omitempty" json:"massdrop,omitempty"`
}

type Massdrop struct {
	Id          bson.ObjectId         `bson:"_id,omitempty" json:"id"`
	ProductId   bson.ObjectId         `bson:"product_id" json:"product_id"`
	Deadline    time.Time             `bson:"deadline" json:"deadline"`
	Price       float64               `bson:"price" json:"price"`
	Reserve     float64               `bson:"reserve_price" json:"reserve_price"`
	Active      bool                  `bson:"active" json:"active"`
	Checkpoints []MassdropCheckpoint  `bson:"checkpoints" json:"checkpoints"`
}

type MassdropCheckpoint struct {
	Step     int     `bson:"step" json:"step"`
	Starts   int     `bson:"starts" json:"starts"`
	Ends     int     `bson:"ends" json:"ends"`
	Price    float64 `bson:"price" json:"price"`
	Timespan int     `bson:"timespan" json:"timespan"`
}

type ProductAggregation struct {
	Category string `bson:"_id" json:"category"`
	Count    int    `bson:"count" json:"count"`
}

func (this *Product) SetDI(i *Module) {
	this.di = i
}