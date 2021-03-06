package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Counter struct {
	UserId   bson.ObjectId          `bson:"user_id" json:"user_id"`
	Counters map[string]PostCounter `bson:"counters" json:"counters"`
}

type PostCounter struct {
	Counter int       `bson:"counter" json:"counter"`
	Updated time.Time `bson:"updated_at" json:"updated_at"`
}

type Notification struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	UserId    bson.ObjectId `bson:"user_id" json:"user_id"`
	RelatedId bson.ObjectId `bson:"related_id" json:"related_id"`
	Title     string        `bson:"title" json:"title"`
	Text      string        `bson:"text" json:"text"`
	Link      string        `bson:"link" json:"link"`
	Related   string        `bson:"related" json:"related"`
	Seen      bool          `bson:"seen" json:"seen"`
	Image     string        `bson:"image" json:"image"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
	Updated   time.Time     `bson:"updated_at" json:"updated_at"`
}

type UserFirebaseNotification struct {
	UserId       bson.ObjectId `json:"user_id"`
	RelatedId    bson.ObjectId `json:"related_id"`
	RelatedExtra string        `bson:"related_extra" json:"related_extra"`
	Position     string        `bson:"position,omitempty" json:"position,omitempty"`
	Title        string        `json:"title,omitempty"`
	Username     string        `json:"username,omitempty"`
	Text         string        `json:"text"`
	Related      string        `json:"related"`
	Seen         bool          `json:"seen"`
	Image        string        `json:"image"`
	Created      time.Time     `json:"created_at"`
	Updated      time.Time     `json:"updated_at"`
}

type UserFirebaseNotifications struct {
	Count int                                 `json:"count"`
	List  map[string]UserFirebaseNotification `json:"list,omitempty"`
}

type MentionModel struct {
	PostId bson.ObjectId `bson:"post_id" json:"post_id"`
	UserId bson.ObjectId `bson:"user_id" json:"user_id"`
	Nested int           `bson:"nested" json:"nested"`
}

type UserFirebase struct {
	Online  int    `json:"online"`
	Viewing string `json:"viewing"`
	Pending int    `json:"pending"`
}
