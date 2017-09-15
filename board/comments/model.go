package comments

import (
	"github.com/fernandez14/spartangeek-blacker/board/votes"
	"gopkg.in/mgo.v2/bson"
)

type Comment struct {
	Id       bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	PostId   bson.ObjectId `bson:"post_id" json:"post_id"`
	UserId   bson.ObjectId `bson:"user_id" json:"user_id"`
	Votes    votes.Votes   `bson:"votes" json:"votes"`
	User     interface{}   `bson:"-" json:"author,omitempty"`
	Position int           `bson:"position" json:"position"`
	Liked    int           `bson:"-" json:"liked,omitempty"`
	Content  string        `bson:"content" json:"content"`
	Chosen   bool          `bson:"chosen,omitempty" json:"chosen,omitempty"`
	Created  time.Time     `bson:"created_at" json:"created_at"`
}
