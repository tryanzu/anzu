package votes

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"time"
)

type voteDir int

const (
	UP voteDir = iota
	DOWN
)

// Vote represents a reaction to a post || comment
type Vote struct {
	ID         bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	UserID     bson.ObjectId `bson:"user_id" json:"user_id"`
	Type       string        `bson:"type" json:"type"`
	NestedType string        `bson:"nested_type,omitempty" json:"nested_type,omitempty"`
	RelatedID  bson.ObjectId `bson:"related_id" json:"related_id"`
	Value      string        `bson:"value" json:"value"`
	Created    time.Time     `bson:"created_at" json:"created_at"`
	Deleted    *time.Time    `bson:"deleted_at,omitempty" json:"-"`
}

func (v Vote) Remove(deps Deps) error {
	return nil
}

func (v Vote) DbField() string {
	return "votes." + v.Value
}

// Votes represents the aggregated count of votes
type Votes map[string]int

/*struct {
	Up     int `bson:"up" json:"up"`
	Down   int `bson:"down" json:"down"`
	Rating int `bson:"rating,omitempty" json:"rating,omitempty"`
}*/

func coll(deps Deps) *mgo.Collection {
	return deps.Mgo().C("votes")
}

// List aggregates a list of votes for certain event.
type List []Vote

func (ls List) ValuesMap() map[string]string {
	m := make(map[string]string, len(ls))
	for _, v := range ls {
		m[v.RelatedID.Hex()] = v.Value
	}
	return m
}
