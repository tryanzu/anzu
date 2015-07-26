package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Votes struct {
	Up     int `bson:"up" json:"up"`
	Down   int `bson:"down" json:"down"`
	Rating int `bson:"rating,omitempty" json:"rating,omitempty"`
}

type Author struct {
	Id      bson.ObjectId `bson:"id,omitempty" json:"id,omitempty"`
	Name    string        `bson:"name" json:"name"`
	Email   string        `bson:"email" json:"email"`
	Avatar  string        `bson:"avatar" json:"avatar"`
	Profile interface{}   `bson:"profile,omitempty" json:"profile,omitempty"`
}

type Comments struct {
	Count int       `bson:"count" json:"count"`
	Set   []Comment `bson:"set" json:"set"`
}

type FeedComments struct {
	Count int       `bson:"count" json:"count"`
}

type Comment struct {
	UserId   bson.ObjectId `bson:"user_id" json:"user_id"`
	Votes    Votes         `bson:"votes" json:"votes"`
	User     interface{}   `bson:"author,omitempty" json:"author,omitempty"`
	Position int           `bson:"position,omitempty" json:"position"`
	Liked    int     	   `bson:"liked,omitempty" json:"liked,omitempty"`
	Content  string        `bson:"content" json:"content"`
	Created  time.Time     `bson:"created_at" json:"created_at"`
}

type Components struct {
	Cpu               Component `bson:"cpu,omitempty" json:"cpu,omitempty"`
	Motherboard       Component `bson:"motherboard,omitempty" json:"motherboard,omitempty"`
	Ram               Component `bson:"ram,omitempty" json:"ram,omitempty"`
	Storage           Component `bson:"storage,omitempty" json:"storage,omitempty"`
	Cooler            Component `bson:"cooler,omitempty" json:"cooler,omitempty"`
	Power             Component `bson:"power,omitempty" json:"power,omitempty"`
	Cabinet           Component `bson:"cabinet,omitempty" json:"cabinet,omitempty"`
	Screen            Component `bson:"screen,omitempty" json:"screen,omitempty"`
	Videocard         Component `bson:"videocard,omitempty" json:"videocard,omitempty"`
	Software          string    `bson:"software,omitempty" json:"software,omitempty"`
	Budget            string    `bson:"budget,omitempty" json:"budget,omitempty"`
	BudgetCurrency    string    `bson:"budget_currency,omitempty" json:"budget_currency,omitempty"`
	BudgetType        string    `bson:"budget_type,omitempty" json:"budget_type,omitempty"`
	BudgetFlexibility string    `bson:"budget_flexibility,omitempty" json:"budget_flexibility,omitempty"`
}

type Component struct {
	Content   string           `bson:"content" json:"content"`
	Elections bool             `bson:"elections" json:"elections"`
	Options   []ElectionOption `bson:"options,omitempty" json:"options"`
	Votes     Votes            `bson:"votes" json:"votes"`
	Status    string           `bson:"status" json:"status"`
	Voted     string           `bson:"voted,omitempty" json:"voted,omitempty"`
}

type Post struct {
	Id         bson.ObjectId   `bson:"_id,omitempty" json:"id,omitempty"`
	Title      string          `bson:"title" json:"title"`
	Slug       string          `bson:"slug" json:"slug"`
	Type       string          `bson:"type" json:"type"`
	Content    string          `bson:"content" json:"content"`
	Categories []string        `bson:"categories" json:"categories"`
	Comments   Comments        `bson:"comments" json:"comments"`
	Author     User            `bson:"author,omitempty" json:"author,omitempty"`
	UserId     bson.ObjectId   `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Users      []bson.ObjectId `bson:"users,omitempty" json:"users,omitempty"`
	Votes      Votes           `bson:"votes" json:"votes"`
	Components Components      `bson:"components,omitempty" json:"components,omitempty"`
	Following  bool            `bson:"following,omitempty" json:"following,omitempty"`
	Pinned     bool            `bson:"pinned,omitempty" json:"pinned,omitempty"`
	NoComments bool            `bson:"comments_blocked" json:"comments_blocked"`
	Created    time.Time       `bson:"created_at" json:"created_at"`
	Updated    time.Time       `bson:"updated_at" json:"updated_at"`
}

type FeedPost struct {
	Id         bson.ObjectId   `bson:"_id,omitempty" json:"id,omitempty"`
	Title      string          `bson:"title" json:"title"`
	Slug       string          `bson:"slug" json:"slug"`
	Type       string          `bson:"type" json:"type"`
	Categories []string        `bson:"categories" json:"categories"`
	Comments   FeedComments     `bson:"comments" json:"comments"`
	Author     User            `bson:"author,omitempty" json:"author,omitempty"`
	UserId     bson.ObjectId   `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Votes      Votes           `bson:"votes" json:"votes"`
	Pinned     bool            `bson:"pinned,omitempty" json:"pinned,omitempty"`
	Created    time.Time       `bson:"created_at" json:"created_at"`
	Updated    time.Time       `bson:"updated_at" json:"updated_at"`
}

type PostForm struct {
	Kind       string                 `json:"kind" binding:"required"`
	Name       string                 `json:"name" binding:"required"`
	Content    string                 `json:"content" binding:"required"`
	Budget     string                 `json:"budget"`
	Currency   string                 `json:"currency"`
	Moves      string                 `json:"moves"`
	Software   string                 `json:"software"`
	Tag        string                 `json:"tag"`
	Components map[string]interface{} `json:"components"`
}

// ByCommentCreatedAt implements sort.Interface for []ElectionOption based on Created field
type ByCommentCreatedAt []Comment

func (a ByCommentCreatedAt) Len() int           { return len(a) }
func (a ByCommentCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCommentCreatedAt) Less(i, j int) bool { return !a[i].Created.Before(a[j].Created) }
