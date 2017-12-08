package comments

import (
	"github.com/fernandez14/spartangeek-blacker/board/votes"
	"github.com/fernandez14/spartangeek-blacker/core/common"
	"github.com/fernandez14/spartangeek-blacker/core/content"
	"github.com/fernandez14/spartangeek-blacker/core/user"
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
	ReplyTo  bson.ObjectId `bson:"reply_to,omitempty" json:"reply_to,omitempty"`
	Chosen   bool          `bson:"chosen,omitempty" json:"chosen,omitempty"`
	Created  time.Time     `bson:"created_at" json:"created_at"`

	// Runtime generated fields.
	Replies interface{} `bson:"-" json:"replies,omitempty"`
}

func (c Comment) GetContent() string {
	return c.Content
}

func (c Comment) UpdateContent(content string) content.Parseable {
	c.Content = content
	return c
}

func (c Comment) GetParseableMeta() (meta map[string]interface{}) {
	return
}

type Replies struct {
	Id    bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Count int           `bson:"count" json:"count"`
	List  Comments      `bson:"list" json:"list"`
}

type RepliesList []Replies

type Comments []Comment

func (all Comments) Map() map[string]Comment {
	m := make(map[string]Comment, len(all))
	for _, item := range all {
		m[item.Id.Hex()] = item
	}

	return m
}

func (all Comments) WithReplies(deps Deps, max int) (Comments, error) {
	list := make(Comments, len(all))
	replies, err := FindReplies(deps, all, max)
	if err != nil {
		return Comments{}, err
	}

	for n, r := range replies {
		replies[n].List, err = r.List.WithUsers(deps)

		if err != nil {
			return Comments{}, err
		}
	}

	for n, c := range all {
		list[n] = c

		for _, r := range replies {
			if r.Id == c.Id {
				list[n].Replies = r
				break
			}
			list[n].Replies = make([]string, 0)
		}
	}

	return list, nil
}

func (all Comments) WithUsers(deps Deps) (Comments, error) {
	list := make(Comments, len(all))
	users, err := user.FindList(deps, all.UsersScope())
	if err != nil {
		return Comments{}, err
	}

	for n, c := range all {
		list[n] = c

		for _, r := range users {
			if r.Id == c.UserId {
				list[n].User = r
				break
			}
		}
	}

	return list, nil
}

func (all Comments) IDList() []bson.ObjectId {
	list := make([]bson.ObjectId, len(all))
	index := 0
	for _, c := range all {
		list[index] = c.Id
		index++
	}

	return list
}

func (all Comments) UsersScope() common.Scope {
	users := map[bson.ObjectId]bool{}
	for _, c := range all {
		if _, exists := users[c.UserId]; !exists {
			users[c.UserId] = true
		}
	}

	list := make([]bson.ObjectId, len(users))
	index := 0
	for k, _ := range users {
		list[index] = k
		index++
	}

	return common.WithinID(list)
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
