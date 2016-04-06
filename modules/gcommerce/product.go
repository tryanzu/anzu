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
	Type        string                 `bson:"type" json:"type"`
	Category    string                 `bson:"category" json:"category"`
	CategoryId  bson.ObjectId          `bson:"category_id,omitempty" json:"category_id,omitempty"`
	Price       float64                `bson:"price" json:"price"`
	Attrs       map[string]interface{} `bson:"attributes" json:"attributes"`
	Created     time.Time              `bson:"created_at" json:"created_at"`
	Updated     time.Time              `bson:"updated_at" json:"updated_at"`

	di *Module
}

func (this *Product) SetDI(i *Module) {
	this.di = i
}