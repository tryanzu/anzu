package notifications

import (
	"github.com/fernandez14/spartangeek-blacker/board/comments"
	posts "github.com/fernandez14/spartangeek-blacker/board/posts"
	"github.com/fernandez14/spartangeek-blacker/core/common"
	"github.com/fernandez14/spartangeek-blacker/core/user"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Notification struct {
	Id        bson.ObjectId   `bson:"_id,omitempty" json:"id,omitempty"`
	UserId    bson.ObjectId   `bson:"user_id" json:"user_id"`
	Type      string          `bson:"type" json:"type"`
	RelatedId bson.ObjectId   `bson:"related_id" json:"related_id"`
	Users     []bson.ObjectId `bson:"users" json:"users"`
	Seen      bool            `bson:"seen" json:"seen"`
	Created   time.Time       `bson:"created_at" json:"created_at"`
	Updated   time.Time       `bson:"updated_at" json:"updated_at"`
}

type Notifications []Notification

func (all Notifications) UsersScope() common.Scope {
	users := map[bson.ObjectId]bool{}
	for _, n := range all {
		users[n.UserId] = true
		for _, id := range n.Users {
			users[id] = true
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

func (all Notifications) CommentsScope() common.Scope {
	comments := map[bson.ObjectId]bool{}
	for _, n := range all {
		if n.Type != "comment" && n.Type != "mention" {
			continue
		}

		comments[n.RelatedId] = true
	}

	list := make([]bson.ObjectId, len(comments))
	index := 0
	for k, _ := range comments {
		list[index] = k
		index++
	}

	return common.WithinID(list)
}

func (all Notifications) Humanize(deps Deps) (list []map[string]interface{}, err error) {
	ulist, err := user.FindList(deps, all.UsersScope())
	if err != nil {
		return
	}

	clist, err := comments.FindList(deps, all.CommentsScope())
	if err != nil {
		return
	}

	plist, err := posts.FindList(deps, clist.PostsScope())
	if err != nil {
		return
	}

	umap := ulist.Map()
	cmap := clist.Map()
	pmap := plist.Map()

	for _, n := range all {
		switch n.Type {
		case "comment":
			comment := cmap[n.RelatedId.Hex()]
			post := pmap[comment.PostId.Hex()]
			user := umap[comment.UserId.Hex()]

			list = append(list, map[string]interface{}{
				"target":    "/p/" + post.Slug + "/" + post.Id.Hex(), /*+ "#c" + comment.Id.Hex()*/
				"title":     "Nuevo comentario de @" + user.UserName,
				"subtitle":  post.Title,
				"createdAt": n.Created,
			})
		case "mention":
			comment := cmap[n.RelatedId.Hex()]
			post := pmap[comment.PostId.Hex()]
			user := umap[comment.UserId.Hex()]

			list = append(list, map[string]interface{}{
				"target":    "/p/" + post.Slug + "/" + post.Id.Hex(), /*+ "#c" + comment.Id.Hex()*/
				"title":     "@" + user.UserName + " te mencionó en un comentario",
				"subtitle":  post.Title,
				"createdAt": n.Created,
			})
		}
	}

	return
}

type Socket struct {
	Chan   string
	Action string
	Params map[string]interface{}
}
