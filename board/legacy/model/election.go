package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type ElectionOption struct {
	UserId  bson.ObjectId `bson:"user_id" json:"user_id"`
	Content string        `bson:"content" json:"content"`
	User    interface{}   `bson:"author,omitempty" json:"author,omitempty"`
	Votes   Votes         `bson:"votes" json:"votes"`
	Created time.Time     `bson:"created_at" json:"created_at"`
}

type ElectionForm struct {
	Component string `json:"component" binding:"required"`
	Content   string `json:"content" binding:"required"`
}

// ByElectionsCreatedAt implements sort.Interface for []ElectionOption based on Created field
type ByElectionsCreatedAt []ElectionOption

func (a ByElectionsCreatedAt) Len() int           { return len(a) }
func (a ByElectionsCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByElectionsCreatedAt) Less(i, j int) bool { return !a[i].Created.Before(a[j].Created) }
