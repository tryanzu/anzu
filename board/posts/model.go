package post

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	Id                primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title             string          `bson:"title" json:"title"`
	Slug              string          `bson:"slug" json:"slug"`
	Type              string          `bson:"type" json:"type"`
	Content           string          `bson:"content" json:"content"`
	Categories        []string        `bson:"categories" json:"categories"`
	Comments          comments        `bson:"comments"`
	Category          primitive.ObjectID `bson:"category" json:"category"`
	UserId            primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Users             []primitive.ObjectID `bson:"users,omitempty" json:"users,omitempty"`
	RelatedComponents []primitive.ObjectID `bson:"related_components,omitempty" json:"related_components,omitempty"`
	Following         bool            `bson:"following,omitempty" json:"following,omitempty"`
	Pinned            bool            `bson:"pinned,omitempty" json:"pinned,omitempty"`
	Lock              bool            `bson:"lock" json:"lock"`
	IsQuestion        bool            `bson:"is_question" json:"is_question"`
	Solved            bool            `bson:"solved,omitempty" json:"solved,omitempty"`
	Liked             int             `bson:"liked,omitempty" json:"liked,omitempty"`
	Created           time.Time       `bson:"created_at" json:"created_at"`
	Updated           time.Time       `bson:"updated_at" json:"updated_at"`
	Deleted           time.Time       `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type comments struct {
	Count int `bson:"count"`
}

func (Post) VotableType() string {
	return "post"
}

func (p Post) VotableID() primitive.ObjectID {
	return p.Id
}

// Posts list.
type Posts []Post

func (list Posts) IDs() []primitive.ObjectID {
	m := make([]primitive.ObjectID, len(list))
	for k, item := range list {
		m[k] = item.Id
	}
	return m
}

func (list Posts) Map() map[primitive.ObjectID]Post {
	m := make(map[primitive.ObjectID]Post, len(list))
	for _, item := range list {
		m[item.Id] = item
	}

	return m
}
