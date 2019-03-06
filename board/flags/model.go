package flags

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type status string
type category string

const (
	PENDING  status = "pending"
	REJECTED status = "rejected"

	SPAM        category = "spam"
	RUDE        category = "rude"
	DUPLICATE   category = "duplicate"
	NEEDSREVIEW category = "needs_review"
)

var categories = []category{SPAM, RUDE, DUPLICATE, NEEDSREVIEW}

// Flag represents a report sent by a user flagging a post/comment.
type Flag struct {
	ID        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    bson.ObjectId `bson:"user_id" json:"user_id"`
	RelatedTo string        `bson:"related_to" json:"related_to"`
	RelatedID bson.ObjectId `bson:"related_id" json:"related_id"`
	Content   string        `bson:"content" json:"content"`
	Status    status        `bson:"status" json:"status"`
	Category  category      `bson:"category" json:"category"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
	Updated   time.Time     `bson:"updated_at" json:"updated_at"`
	Deleted   *time.Time    `bson:"deleted_at,omitempty" json:"-"`
}

// IsValidCategory for a flag.
func IsValidCategory(s string) bool {
	for _, c := range categories {
		if string(c) == s {
			return true
		}
	}
	return false
}

func CastCategory(s string) (category, error) {
	if v := IsValidCategory(s); v == false {
		return "", errors.New("invalid category")
	}
	return category(s), nil
}
