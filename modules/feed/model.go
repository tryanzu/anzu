package feed

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LightPostModel struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title      string             `bson:"title" json:"title"`
	Slug       string             `bson:"slug" json:"slug"`
	Content    string             `bson:"content" json:"content"`
	Type       string             `bson:"type" json:"type"`
	Category   primitive.ObjectID `bson:"category" json:"category"`
	UserId     primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Pinned     bool          `bson:"pinned,omitempty" json:"pinned,omitempty"`
	IsQuestion bool          `bson:"is_question,omitempty" json:"is_question"`
	Solved     bool          `bson:"solved,omitempty" json:"solved,omitempty"`
	Lock       bool          `bson:"lock" json:"lock"`
	BestAnswer *Comment      `bson:"-" json:"best_answer,omitempty"`
	Created    time.Time     `bson:"created_at" json:"created_at"`
	Updated    time.Time     `bson:"updated_at" json:"updated_at"`
}

type PostCommentModel struct {
	Id      primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Comment Comment       `bson:"comment" json:"comment,omitempty"`
}

type PostCommentCountModel struct {
	Id    primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Count int           `bson:"count" json:"count"`
}

type VotesModel struct {
	Up     int `bson:"up" json:"up"`
	Down   int `bson:"down" json:"down"`
	Rating int `bson:"rating,omitempty" json:"rating,omitempty"`
}
