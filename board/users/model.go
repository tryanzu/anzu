package users

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type status string
type category string

const (
	ACTIVE   status = "active"
	PENDING  status = "pending"
	REJECTED status = "rejected"
	REVOKED  status = "revoked"
)

// Ban represents a ban sent by a user.
type Ban struct {
	ID        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    bson.ObjectId `bson:"user_id" json:"user_id"`
	RelatedTo string        `bson:"related_to" json:"related_to"`
	RelatedID bson.ObjectId `bson:"related_id" json:"related_id"`
	Content   string        `bson:"content" json:"content"`
	Status    status        `bson:"status" json:"status"`
	Reason    string        `bson:"reason" json:"reason"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
	Updated   time.Time     `bson:"updated_at" json:"updated_at"`
	Deleted   *time.Time    `bson:"deleted_at,omitempty" json:"-"`
}
