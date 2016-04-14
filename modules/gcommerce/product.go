package gcommerce

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Product struct {
	Id          bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name"`
	Description string                 `bson:"description" json:"description"`
	Slug        string                 `bson:"slug" json:"slug"`
	Image       string                 `bson:"image" json:"image"`
	Images      []string               `bson:"images" json:"images"`
	Type        string                 `bson:"type" json:"type"`
	Category    string                 `bson:"category" json:"category"`
	CategoryId  bson.ObjectId          `bson:"category_id,omitempty" json:"category_id,omitempty"`
	Price       float64                `bson:"price" json:"price"`
	Shipping    float64                `bson:"shipping_cost" json:"shipping_cost"`
	Search      string                 `bson:"search" json:"search"`
	Relevance   int                    `bson:"relevance" json:"relevance"`
	Attrs       map[string]interface{} `bson:"attributes" json:"attributes"`
	Created     time.Time              `bson:"created_at" json:"created_at"`
	Updated     time.Time              `bson:"updated_at" json:"updated_at"`

	di       *Module
	Massdrop *Massdrop `bson:"massdrop,omitempty" json:"massdrop,omitempty"`
}

const MASSDROP_TRANS_RESERVATION = "reservation"
const MASSDROP_TRANS_INSTERESTED = "interested"
const MASSDROP_STATUS_COMPLETED = "completed"
const MASSDROP_STATUS_REMOVED = "removed"

type Massdrop struct {
	Id          bson.ObjectId        `bson:"_id,omitempty" json:"id"`
	ProductId   bson.ObjectId        `bson:"product_id" json:"product_id"`
	Deadline    time.Time            `bson:"deadline" json:"deadline"`
	Price       float64              `bson:"price" json:"price"`
	Reserve     float64              `bson:"reserve_price" json:"reserve_price"`
	Active      bool                 `bson:"active" json:"active"`
	Checkpoints []MassdropCheckpoint `bson:"checkpoints" json:"checkpoints"`

	// Runtime generated data
	Activities   []MassdropActivity `bson:"-" json:"activities"`
	Reservations int                `bson:"-" json:"count_reservations"`
	Interested   int                `bson:"-" json:"count_interested"`
	Current      string             `bson:"-" json:"current,omitempty"`
}

type MassdropCheckpoint struct {
	Step     int     `bson:"step" json:"step"`
	Starts   int     `bson:"starts" json:"starts"`
	Ends     int     `bson:"ends" json:"ends"`
	Price    float64 `bson:"price" json:"price"`
	Timespan int     `bson:"timespan" json:"timespan"`
	Done     bool    `bson:"-" json:"done"`
}

type MassdropActivity struct {
	Type    string                 `bson:"type" json:"type"`
	Attrs   map[string]interface{} `bson:"attributes" json:"attributes"`
	Created time.Time              `bson:"created_at" json:"created_at"`
}

// ByCommentCreatedAt implements sort.Interface for []ElectionOption based on Created field
type MassdropByCreated []MassdropActivity

func (a MassdropByCreated) Len() int           { return len(a) }
func (a MassdropByCreated) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a MassdropByCreated) Less(i, j int) bool { return !a[i].Created.Before(a[j].Created) }

type MassdropTransaction struct {
	Id         bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	MassdropId bson.ObjectId          `bson:"massdrop_id" json:"massdrop_id"`
	CustomerId bson.ObjectId          `bson:"customer_id" json:"customer_id"`
	Type       string                 `bson:"type" json:"type"`
	Status     string                 `bson:"status" json:"status"`
	Attrs      map[string]interface{} `bson:"attributes" json:"attributes"`
	Created    time.Time              `bson:"created_at" json:"created_at"`
	Updated    time.Time              `bson:"updated_at" json:"updated_at"`

	di *Module
}

type ProductAggregation struct {
	Category string `bson:"_id" json:"category"`
	Count    int    `bson:"count" json:"count"`
}

func (this *Product) SetDI(i *Module) {
	this.di = i
}
