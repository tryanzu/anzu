package feed

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type AlgoliaPostModel struct {
	Id         string               `json:"objectID"`
	Title      string               `json:"title"`
	Content    string               `json:"content"`
	Slug       string               `json:"slug"`
	Comments   int                  `json:"comments_count"`
	User       AlgoliaUserModel     `json:"user"`
	Tribute    int                  `json:"tribute_count"`
	Shit       int                  `json:"shit_count"`
	Category   AlgoliaCategoryModel `json:"category"`
	Popularity float64              `json:"popularity"`
	Components []string             `json:"components,omitempty"`
	Created    int64                `json:"created"`
}

type AlgoliaCategoryModel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AlgoliaUserModel struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Image    string `json:"image"`
	Email    string `json:"email"`
}

type LightPostModel struct {
	Id         bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Title      string        `bson:"title" json:"title"`
	Slug       string        `bson:"slug" json:"slug"`
	Content    string        `bson:"content" json:"content"`
	Type       string        `bson:"type" json:"type"`
	Category   bson.ObjectId `bson:"category" json:"category"`
	UserId     bson.ObjectId `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Pinned     bool          `bson:"pinned,omitempty" json:"pinned,omitempty"`
	IsQuestion bool          `bson:"is_question,omitempty" json:"is_question,omitempty"`
	Created    time.Time     `bson:"created_at" json:"created_at"`
	Updated    time.Time     `bson:"updated_at" json:"updated_at"`
}