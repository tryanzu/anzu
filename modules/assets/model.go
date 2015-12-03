package assets

import(
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Asset struct {
	Id bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Related string `bson:"related" json:"related"`
	RelatedId bson.ObjectId `bson:"related_id" json:"related_id"`
	Path string `bson:"path" json:"path"`
	Meta interface{} `bson:"meta" json:"meta"`
	Created  time.Time `bson:"created_at" json:"created_at"`
}