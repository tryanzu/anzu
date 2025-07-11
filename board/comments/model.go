package comments

import (
	"github.com/tryanzu/core/board/votes"
	"github.com/tryanzu/core/core/common"
	"github.com/tryanzu/core/core/content"
	"github.com/tryanzu/core/core/user"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"time"
)

type Comment struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserId    primitive.ObjectID `bson:"user_id" json:"user_id"`
	PostId    primitive.ObjectID `bson:"post_id,omitempty" json:"post_id,omitempty"`
	Votes     votes.Votes   `bson:"votes" json:"votes"`
	User      interface{}   `bson:"-" json:"author,omitempty"`
	Position  int           `bson:"position" json:"-"`
	Liked     int           `bson:"-" json:"liked,omitempty"`
	Content   string        `bson:"content" json:"content"`
	ReplyTo   primitive.ObjectID `bson:"reply_to,omitempty" json:"reply_to,omitempty"`
	ReplyType string        `bson:"reply_type,omitempty" json:"reply_type,omitempty"`
	Chosen    bool          `bson:"chosen,omitempty" json:"chosen,omitempty"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
	Updated   time.Time     `bson:"updated_at" json:"updated_at"`
	Deleted   *time.Time    `bson:"deleted_at,omitempty" json:"-"`

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

func (c Comment) GetParseableMeta() map[string]interface{} {
	meta := make(map[string]interface{})
	meta["id"] = c.Id
	meta["related"] = "comment"
	meta["user_id"] = c.UserId
	return meta
}

func (c Comment) RelatedID() primitive.ObjectID {
	return c.ReplyTo
}

func (c Comment) RelatedPost() primitive.ObjectID {
	if c.ReplyType == "post" {
		return c.ReplyTo
	}
	return c.PostId
}

func (c Comment) VotableType() string {
	return "comment"
}

func (c Comment) VotableID() primitive.ObjectID {
	return c.Id
}

type Replies struct {
	Id    primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	Count int           `bson:"count" json:"count"`
	List  Comments      `bson:"list" json:"list"`
}

type RepliesList []Replies

type CommentsSet struct {
	List  Comments `json:"list"`
	Count int      `json:"count"`
}

type Comments []Comment

func (all Comments) Map() map[primitive.ObjectID]Comment {
	m := make(map[primitive.ObjectID]Comment, len(all.NestedIDList()))
	for _, c := range all {
		m[c.Id] = c

		if c.Replies != nil {
			for _, r := range c.Replies.(Replies).List {
				m[r.Id] = r
			}
		}
	}

	return m
}

func (all Comments) StrMap() map[string]Comment {
	m := make(map[string]Comment, len(all.NestedIDList()))
	for _, c := range all {
		m[c.Id.Hex()] = c

		if c.Replies != nil {
			for _, r := range c.Replies.(Replies).List {
				m[r.Id.Hex()] = r
			}
		}
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
		var processed content.Parseable
		for cindex, c := range replies[n].List {
			processed, err = content.Postprocess(deps, c)
			if err != nil {
				return Comments{}, err
			}

			replies[n].List[cindex] = processed.(Comment)
		}

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
		}
	}

	return list, nil
}

// WithUsers nested data loaded.
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

func (all Comments) NestedIDList() (list []primitive.ObjectID) {
	for _, c := range all {
		list = append(list, c.Id)
		if c.Replies != nil {
			for _, r := range c.Replies.(Replies).List {
				list = append(list, r.Id)
			}
		}
	}
	return
}

func (all Comments) IDList() []primitive.ObjectID {
	list := make([]primitive.ObjectID, len(all))
	index := 0
	for _, c := range all {
		list[index] = c.Id
		index++
	}

	for i := len(list)/2 - 1; i >= 0; i-- {
		opp := len(list) - 1 - i
		list[i], list[opp] = list[opp], list[i]
	}

	return list
}

func (all Comments) UsersScope() common.Scope {
	users := map[primitive.ObjectID]struct{}{}
	for _, c := range all {
		if _, exists := users[c.UserId]; !exists {
			users[c.UserId] = struct{}{}
		}
	}

	list := make([]primitive.ObjectID, len(users))
	index := 0
	for k := range users {
		list[index] = k
		index++
	}

	return common.WithinID(list)
}

func (all Comments) PostIDs() []primitive.ObjectID {
	posts := map[primitive.ObjectID]struct{}{}
	for _, c := range all {
		if !c.PostId.IsZero() {
			posts[c.PostId] = struct{}{}
		}
		if _, exists := posts[c.ReplyTo]; c.ReplyType == "post" && !exists {
			posts[c.ReplyTo] = struct{}{}
		}
	}

	list := make([]primitive.ObjectID, len(posts))
	index := 0
	for k := range posts {
		list[index] = k
		index++
	}
	return list
}

func (all Comments) PostsScope() common.Scope {
	return common.WithinID(all.PostIDs())
}

// VotesOf userId in comments resultset.
func (all Comments) VotesOf(deps Deps, userID primitive.ObjectID) (list votes.List, err error) {
	list, err = votes.FindList(deps, common.WithinID(all.NestedIDList()))
	return
}
