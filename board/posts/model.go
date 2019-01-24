package post

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Post struct {
	Id                bson.ObjectId   `bson:"_id,omitempty" json:"id,omitempty"`
	Title             string          `bson:"title" json:"title"`
	Slug              string          `bson:"slug" json:"slug"`
	Type              string          `bson:"type" json:"type"`
	Content           string          `bson:"content" json:"content"`
	Categories        []string        `bson:"categories" json:"categories"`
	Category          bson.ObjectId   `bson:"category" json:"category"`
	UserId            bson.ObjectId   `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Users             []bson.ObjectId `bson:"users,omitempty" json:"users,omitempty"`
	RelatedComponents []bson.ObjectId `bson:"related_components,omitempty" json:"related_components,omitempty"`
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

func (Post) VotableType() string {
	return "post"
}

func (p Post) VotableID() bson.ObjectId {
	return p.Id
}

// Posts list.
type Posts []Post

func (list Posts) IDs() []bson.ObjectId {
	m := make([]bson.ObjectId, len(list))
	for k, item := range list {
		m[k] = item.Id
	}
	return m
}

func (list Posts) Map() map[bson.ObjectId]Post {
	m := make(map[bson.ObjectId]Post, len(list))
	for _, item := range list {
		m[item.Id] = item
	}

	return m
}
