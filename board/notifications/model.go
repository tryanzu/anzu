package notifications

import (
	"time"

	"github.com/tryanzu/core/board/comments"
	posts "github.com/tryanzu/core/board/posts"
	"github.com/tryanzu/core/core/common"
	"github.com/tryanzu/core/core/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	Id        primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	UserId    primitive.ObjectID   `bson:"user_id" json:"user_id"`
	Type      string               `bson:"type" json:"type"`
	RelatedId primitive.ObjectID   `bson:"related_id" json:"related_id"`
	Users     []primitive.ObjectID `bson:"users" json:"users"`
	Seen      bool                 `bson:"seen" json:"seen"`
	Created   time.Time            `bson:"created_at" json:"created_at"`
	Updated   time.Time            `bson:"updated_at" json:"updated_at"`
}

type Notifications []Notification

func (all Notifications) UsersScope() common.Scope {
	users := map[primitive.ObjectID]bool{}
	for _, n := range all {
		users[n.UserId] = true
		for _, id := range n.Users {
			users[id] = true
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

func (all Notifications) CommentsScope() common.Scope {
	comments := map[primitive.ObjectID]struct{}{}
	for _, n := range all {
		if n.Type != "comment" && n.Type != "mention" {
			continue
		}

		comments[n.RelatedId] = struct{}{}
	}

	list := make([]primitive.ObjectID, len(comments))
	index := 0
	for k := range comments {
		list[index] = k
		index++
	}

	return common.WithinID(list)
}

func (all Notifications) Humanize(deps Deps) (list []map[string]interface{}, err error) {
	ulist, err := user.FindList(deps, all.UsersScope())
	if err != nil {
		panic(err)
		return
	}

	clist, err := comments.FindList(deps, all.CommentsScope())
	if err != nil {
		panic(err)
		return
	}

	plist, err := posts.FindList(deps, clist.PostsScope())
	if err != nil {
		panic(err)
		return
	}

	umap := ulist.Map()
	cmap := clist.Map()
	pmap := plist.Map()

	for _, n := range all {
		switch n.Type {
		case "comment":
			comment := cmap[n.RelatedId]
			post := pmap[comment.RelatedPost()]
			user := umap[comment.UserId]

			list = append(list, map[string]interface{}{
				"id":        n.Id.Hex(),
				"target":    "/p/" + post.Slug + "/" + post.Id.Hex() + "#" + n.RelatedId.Hex(),
				"title":     "Nuevo comentario de @" + user.UserName,
				"subtitle":  post.Title,
				"createdAt": n.Created,
			})
		case "mention":
			comment := cmap[n.RelatedId]
			post := pmap[comment.ReplyTo]
			user := umap[comment.UserId]

			list = append(list, map[string]interface{}{
				"id":        n.Id.Hex(),
				"target":    "/p/" + post.Slug + "/" + post.Id.Hex(), /*+ "#c" + comment.Id.Hex()*/
				"title":     "@" + user.UserName + " te mencionó en un comentario",
				"subtitle":  post.Title,
				"createdAt": n.Created,
			})
		case "chat":
			user := umap[n.Users[0]]
			list = append(list, map[string]interface{}{
				"id":        n.Id.Hex(),
				"target":    "/chat",
				"title":     "@" + user.UserName + " te mencionó en el chat",
				"createdAt": n.Created,
			})
		}
	}

	return
}

type Socket struct {
	Chan   string                 `json:"c"`
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"p"`
}
