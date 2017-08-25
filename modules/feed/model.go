package feed

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type LightPostModel struct {
	Id         bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Title      string        `bson:"title" json:"title"`
	Slug       string        `bson:"slug" json:"slug"`
	Content    string        `bson:"content" json:"content"`
	Type       string        `bson:"type" json:"type"`
	Category   bson.ObjectId `bson:"category" json:"category"`
	UserId     bson.ObjectId `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Pinned     bool          `bson:"pinned,omitempty" json:"pinned,omitempty"`
	IsQuestion bool          `bson:"is_question,omitempty" json:"is_question"`
	Solved     bool          `bson:"solved,omitempty" json:"solved,omitempty"`
	Lock       bool          `bson:"lock" json:"lock"`
	BestAnswer *Comment      `bson:"-" json:"best_answer,omitempty"`
	Created    time.Time     `bson:"created_at" json:"created_at"`
	Updated    time.Time     `bson:"updated_at" json:"updated_at"`
}

type SearchPostModel struct {
	Id         bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Title      string        `bson:"title" json:"title"`
	Slug       string        `bson:"slug" json:"slug"`
	Content    string        `bson:"content" json:"content"`
	Type       string        `bson:"type" json:"type"`
	Category   bson.ObjectId `bson:"category" json:"category"`
	UserId     bson.ObjectId `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Pinned     bool          `bson:"pinned,omitempty" json:"pinned,omitempty"`
	IsQuestion bool          `bson:"is_question,omitempty" json:"is_question,omitempty"`
	Solved     bool          `bson:"solved,omitempty" json:"solved,omitempty"`
	Lock       bool          `bson:"lock" json:"lock"`
	Score      float64       `bson:"score" json:"score"`
	User       interface{}   `bson:"-" json:"user,omitempty"`
	Created    time.Time     `bson:"created_at" json:"created_at"`
	Updated    time.Time     `bson:"updated_at" json:"updated_at"`
}

type PostCommentModel struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Comment Comment       `bson:"comment" json:"comment,omitempty"`
}

type PostCommentCountModel struct {
	Id    bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Count int           `bson:"count" json:"count"`
}

type VotesModel struct {
	Up     int `bson:"up" json:"up"`
	Down   int `bson:"down" json:"down"`
	Rating int `bson:"rating,omitempty" json:"rating,omitempty"`
}
