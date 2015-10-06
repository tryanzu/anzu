package store

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type OrderModel struct {
	Id       bson.ObjectId  `bson:"_id,omitempty" json:"id,omitempty"`
	User     OrderUserModel `bson:"user" json:"user"`
	Content  string         `bson:"content" json:"content"`
	Budget   int            `bson:"budget" json:"budget"`
	Currency string         `bson:"currency" json:"currency"`
	Games    []string       `bson:"games" json:"games"`
	Extra    []string       `bson:"extras" json:"extra"`
	BuyDelay int            `bson:"buydelay" json:"buydelay"`
	Created  time.Time      `bson:"created_at" json:"created_at"`
	Updated  time.Time      `bson:"updated_at" json:"updated_at"`
}

type OrderUserModel struct {
	Name  string `bson:"name" json:"name"`
	Email string `bson:"email" json:"email"`
	Phone string `bson:"phone" json:"phone"`
}

type MessageModel struct {
	Type      string        `bson:"type" json:"type"`
	Content   string        `bson:"content" json:"content"`
	RelatedId bson.ObjectId `bson:"related_id,omitempty" json:"related_id,omitempty"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
	Updated   time.Time     `bson:"updated_at" json:"updated_at"`
}
