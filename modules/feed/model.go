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
	IsQuestion bool          `bson:"is_question,omitempty" json:"is_question"`
	Solved     bool          `bson:"solved,omitempty" json:"solved,omitempty"`
	Lock       bool          `bson:"lock" json:"lock"`
	BestAnswer *CommentModel `json:"best_answer,omitempty"`
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

type CommentModel struct {
	UserId  bson.ObjectId `bson:"user_id" json:"user_id"`
	Votes   VotesModel    `bson:"votes" json:"votes"`
	Content string        `bson:"content" json:"content"`
	Chosen  bool          `bson:"chosen,omitempty" json:"chosen,omitempty"`
	Created time.Time     `bson:"created_at" json:"created_at"`
	Deleted time.Time     `bson:"deleted_at" json:"deleted_at"`
}

type PostCommentModel struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Comment CommentModel  `bson:"comment" json:"comment,omitempty"`
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
