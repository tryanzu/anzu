package gcommerce

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Customer struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	UserId  bson.ObjectId `bson:"user_id" json:"user_id"`
	Created time.Time     `bson:"created_at" json:"created_at"`
	Updated time.Time     `bson:"updated_at" json:"updated_at"`

	di *Module

	Addresses []CustomerAddress `bson:"-" json:"addresses"`
}

type CustomerAddress struct {
	Id         bson.ObjectId    `bson:"_id,omitempty" json:"id,omitempty"`
	CustomerId bson.ObjectId    `bson:"customer_id" json:"customer_id"`
	Alias      string           `bson:"alias" json:"alias"`
	Slug       string           `bson:"slug" json:"slug"`
	Recipient  string           `bson:"recipient" json:"recipient"`
	Phone      string           `bson:"phone" json:"phone"`
	
	Address    Address          `bson:"address" json:"address"`
	TimesUsed  int              `bson:"times_used" json:"times_used"`
	LastUsed   time.Time        `bson:"last_used" json:"last_used"`
	Default    bool             `bson:"default" json:"default"`
	Created    time.Time        `bson:"created_at" json:"created_at"`
	Updated    time.Time        `bson:"updated_at" json:"updated_at"`

	di *Module
}