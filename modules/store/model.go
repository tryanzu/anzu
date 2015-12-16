package store

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type OrderModel struct {
	Id         bson.ObjectId   `bson:"_id,omitempty" json:"id,omitempty"`
	User       OrderUserModel  `bson:"user" json:"user"`
	Content    string          `bson:"content" json:"content"`
	Budget     int             `bson:"budget" json:"budget"`
	Currency   string          `bson:"currency" json:"currency"`
	State      string          `bson:"state" json:"state"`
	Usage      string          `bson:"usage" json:"usage"`
	Games      []string        `bson:"games" json:"games"`
	Extra      []string        `bson:"extras" json:"extra"`
	BuyDelay   int             `bson:"buydelay" json:"buydelay"`
	Unreaded   bool            `bson:"unreaded" json:"unreaded"`
	Messages   []MessageModel  `bson:"messages,omitempty" json:"messages"`
	Tags       []TagModel      `bson:"tags,omitempty" json:"tags"`
	Activities []ActivityModel `bson:"activities,omitempty" json:"activities"`
	Pipeline   PipelineModel   `bson:"pipeline,omitempty" json:"pipeline"`
	Created    time.Time       `bson:"created_at" json:"created_at"`
	Updated   time.Time        `bson:"updated_at" json:"updated_at"`

	// Runtime generated and not persisted in database
	RelatedUsers interface{} `bson:"-" json:"related_users,omitempty"`
}

type OrderUserModel struct {
	Name  string `bson:"name" json:"name"`
	Email string `bson:"email" json:"email"`
	Phone string `bson:"phone" json:"phone"`
	Ip    string `bson:"ip" json:"ip"`
}

type MessageModel struct {
	Type      string        `bson:"type" json:"type"`
	Content   string        `bson:"content" json:"content"`
	RelatedId bson.ObjectId `bson:"related_id,omitempty" json:"related_id,omitempty"`
	Meta      map[string]interface{} `bson:"-" json:"meta"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
	Updated   time.Time     `bson:"updated_at" json:"updated_at"`
}

type ActivityModel struct {
	Name        string    `bson:"name" json:"name"`
	Description string    `bson:"description" json:"description"`
	Done        bool      `bson:"done" json:"done"`
	Due         time.Time `bson:"due_at" json:"due_at"`
	Created     time.Time `bson:"created_at" json:"created_at"`
	Updated     time.Time `bson:"updated_at" json:"updated_at"`
}

type TagModel struct {
	Name    string        `bson:"name" json:"name"`
	Created time.Time     `bson:"created_at" json:"created_at"`
}

type PipelineModel struct {
	Current string                `bson:"current" json:"current"`
	Step    int                   `bson:"step" json:"step"`
	Updated time.Time             `bson:"updated_at" json:"updated_at"`
	Changes []PipelineChangeModel `bson:"changes" json:"changes"`
}

type PipelineChangeModel struct {
	Name    string    `bson:"name" json:"name"`
	Diff    int       `bson:"diff" json:"diff"`
	Updated time.Time `bson:"updated_at" json:"updated_at"`
}

type BuildResponseModel struct {
	Id      bson.ObjectId  `bson:"_id,omitempty" json:"id,omitempty"`
	Title   string         `bson:"title" json:"title"`
	Content string         `bson:"content" json:"content"`
	Price   int            `bson:"price,omitempty" json:"price,omitempty"`
}

type ByCreatedAt []MessageModel

func (a ByCreatedAt) Len() int           { return len(a) }
func (a ByCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreatedAt) Less(i, j int) bool { return !a[i].Created.Before(a[j].Created) }