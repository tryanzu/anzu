package flags

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type status string

const (
	PENDING  status = "pending"
	REJECTED status = "rejected"
)

// Flag represents a report sent by a user flagging a post/comment.
type Flag struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID  `bson:"user_id" json:"user_id"`
	RelatedTo string              `bson:"related_to" json:"related_to"`
	RelatedID *primitive.ObjectID `bson:"related_id" json:"related_id,omitempty"`
	Content   string              `bson:"content" json:"content"`
	Status    status              `bson:"status" json:"status"`
	Reason    string              `bson:"reason" json:"reason"`
	Created   time.Time           `bson:"created_at" json:"created_at"`
	Updated   time.Time           `bson:"updated_at" json:"updated_at"`
	Deleted   *time.Time          `bson:"deleted_at,omitempty" json:"-"`
}
