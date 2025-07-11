package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Activity struct {
	Id        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserId    primitive.ObjectID   `bson:"user_id,omitempty" json:"user_id"`
	Event     string               `bson:"event,omitempty" event:"related"`
	RelatedId primitive.ObjectID   `bson:"related_id,omitempty" json:"related_id,omitempty"`
	List      []primitive.ObjectID `bson:"list,omitempty" json:"list,omitempty"`
	Created   time.Time            `bson:"created_at" json:"created_at"`
}
