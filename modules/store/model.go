package store

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Invoice struct {
	Id      bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	DealId  bson.ObjectId          `bson:"deal_id,omitempty" json:"deal_id,omitempty"`
	Assets  InvoiceAssets          `bson:"assets" json:"assets"`
	Meta    map[string]interface{} `bson:"meta" json:"meta"`
	Created time.Time              `bson:"created_at" json:"created_at"`
	Updated time.Time              `bson:"updated_at" json:"updated_at"`
}

type InvoiceAssets struct {
	XML string `bson:"xml" json:"xml"`
	PDF string `bson:"pdf" json:"pdf"`
}

type Messages []MessageModel

func (a Messages) Len() int           { return len(a) }
func (a Messages) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Messages) Less(i, j int) bool { return !a[i].Created.Before(a[j].Created) }

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
	Messages   Messages        `bson:"messages,omitempty" json:"messages"`
	Tags       []TagModel      `bson:"tags,omitempty" json:"tags"`
	Activities []ActivityModel `bson:"activities,omitempty" json:"activities"`
	Pipeline   PipelineModel   `bson:"pipeline,omitempty" json:"pipeline"`
	Trusted    bool            `bson:"trusted_flag" json:"trusted_flag"`
	Favorite   bool            `bson:"favorite_flag" json:"favorite_flag"`
	Lead       bool            `bson:"-" json:"lead"`
	Created    time.Time       `bson:"created_at" json:"created_at"`
	Updated    time.Time       `bson:"updated_at" json:"updated_at"`

	// Runtime generated and not persisted in database
	RelatedUsers interface{}  `bson:"-" json:"related_users,omitempty"`
	Duplicates   []OrderModel `bson:"-" json:"duplicates,omitempty"`
	Invoice      *Invoice     `bson:"-" json:"invoice,omitempty"`
}

type OrderUserModel struct {
	Name  string `bson:"name" json:"name"`
	Email string `bson:"email" json:"email"`
	Phone string `bson:"phone" json:"phone"`
	Ip    string `bson:"ip" json:"ip"`
}

type MessageModel struct {
	Type        string                 `bson:"type" json:"type"`
	Content     string                 `bson:"content" json:"content"`
	MessageID   string                 `bson:"postmark_id,omitempty" json:"postmark_id,omitempty"`
	RelatedId   bson.ObjectId          `bson:"related_id,omitempty" json:"related_id,omitempty"`
	TrackId     bson.ObjectId          `bson:"otrack_id,omitempty" json:"otrack_id,omitempty"`
	Opened      time.Time              `bson:"opened_at,omitempty" json:"opened_at,omitempty"`
	ReadSeconds int                    `bson:"read_seconds,omitempty" json:"read_seconds,omitempty"`
	Meta        map[string]interface{} `bson:"-" json:"meta,omitempty"`
	Created     time.Time              `bson:"created_at" json:"created_at"`
	Updated     time.Time              `bson:"updated_at" json:"updated_at"`
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
	Name    string    `bson:"name" json:"name"`
	Created time.Time `bson:"created_at" json:"created_at"`
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
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Title   string        `bson:"title" json:"title"`
	Content string        `bson:"content" json:"content"`
	Price   int           `bson:"price,omitempty" json:"price,omitempty"`
}

type ByCreatedAt []MessageModel

func (a ByCreatedAt) Len() int           { return len(a) }
func (a ByCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreatedAt) Less(i, j int) bool { return !a[i].Created.Before(a[j].Created) }
