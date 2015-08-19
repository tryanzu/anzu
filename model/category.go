package model

import (
	"gopkg.in/mgo.v2/bson"
)

type Category struct {
	Id          bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	Slug        string        `bson:"slug" json:"slug"`
	Color       string        `bson:"color" json:"color"`
	Permissions interface{}   `bson:"permissions" json:"permissions"`
	Parent      bson.ObjectId `bson:"parent,omitempty" json:"parent,omitempty"`
	Count       int           `bson:"count,omitempty" json:"count,omitempty"`
	Recent      int           `bson:"recent,omitempty" json:"recent,omitempty"`
	Child		[]Category    `bson:"subcategories,omitempty" json:"subcategories,omitempty"`
}

type CategoryCounters struct {
	List []CategoryCounter `json:"list"`
}

type CategoryCounter struct {
	Slug  string `json:"slug"`
	Count int    `json:"count"`
}

type Categories []Category

func (slice Categories) Len() int {
	return len(slice)
}

func (slice Categories) Less(i, j int) bool {
	return slice[i].Count > slice[j].Count
}

func (slice Categories) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
