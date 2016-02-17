package cli

import (
	"gopkg.in/mgo.v2/bson"
)

type PostModel struct {
	Id         string        `json:"objectID"`
	Title      string        `json:"title"`
	Content    string        `json:"content"`
	Comments   int           `json:"comments_count"`
	User       UserModel     `json:"user"`
	Tribute    int           `json:"tribute_count"`
	Shit       int           `json:"shit_count"`
	Category   CategoryModel `json:"category"`
	Popularity float64       `json:"popularity"`
	Components []string      `json:"components,omitempty"`
	Created    int64         `json:"created"`
}

type CategoryModel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type NewsletterModel struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Value   NewsletterValueModel        `bson:"value" json:"value"`
}

type NewsletterValueModel struct {
	Email string `bson:"email" json:"email"`
}

type UserModel struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Image    string `json:"image"`
	Email    string `json:"email"`
}
