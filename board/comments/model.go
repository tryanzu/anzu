package comments

import (
	"github.com/fernandez14/spartangeek-blacker/board/votes"
	"github.com/fernandez14/spartangeek-blacker/core/common"
	"gopkg.in/mgo.v2/bson"

	"time"
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

type Comments []Comment

func (list Comments) Map() map[string]Comment {
	m := make(map[string]Comment, len(list))
	for _, item := range list {
		m[item.Id.Hex()] = item
	}

	return m
}

func (all Comments) PostsScope() common.Scope {
	posts := map[bson.ObjectId]bool{}
	for _, c := range all {
		if _, exists := posts[c.PostId]; !exists {
			posts[c.PostId] = true
		}
	}

	list := make([]bson.ObjectId, len(posts))
	index := 0
	for k, _ := range posts {
		list[index] = k
		index++
	}

	return common.WithinID(list)
}
