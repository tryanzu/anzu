package activity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// M stands for activity model.
type M struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id,omitempty" json:"user_id"`
	Event     string          `bson:"event,omitempty" event:"related"`
	RelatedID primitive.ObjectID `bson:"related_id,omitempty" json:"related_id,omitempty"`
	List      []primitive.ObjectID `bson:"list,omitempty" json:"list,omitempty"`
	Created   time.Time       `bson:"created_at" json:"created_at"`
}
