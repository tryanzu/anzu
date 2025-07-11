package assets

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Asset struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Related   string             `bson:"related" json:"related"`
	RelatedId primitive.ObjectID `bson:"related_id" json:"related_id"`
	Path      string             `bson:"path" json:"path"`
	Meta      interface{}        `bson:"meta" json:"meta"`
	Created   time.Time          `bson:"created_at" json:"created_at"`
}
