package events

import (
	"fmt"

	"github.com/tryanzu/core/board/comments"
	notify "github.com/tryanzu/core/board/notifications"
	"github.com/tryanzu/core/board/posts"
	"github.com/tryanzu/core/board/votes"
	pool "github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/gaming"
	"gopkg.in/mgo.v2/bson"
)

// Bind event handlers for comment related actions...
func commentsEvents() {
	handlers := []pool.EventHandler{
		{
			On:      pool.POSTS_COMMENT,
			Handler: onPostComment,
		},
		{
			On:      pool.COMMENT_DELETE,
			Handler: onCommentDelete,
		},
		{
			On:      pool.COMMENT_UPDATE,
			Handler: onCommentUpdate,
		},
		{
			On:      pool.VOTE,
			Handler: onVote,
		},
	}

	register(handlers)
}

func onVote(e pool.Event) error {
	var (
		err    error
		userID bson.ObjectId
	)
	vote := e.Params["vote"].(votes.Vote)
	field := vote.DbField()

	// Increment value by given state.
	value := 1
	if vote.Deleted != nil {
		value = -1
	}
	switch vote.Type {
	case "comment":
		err = deps.Container.Mgo().C("comments").UpdateId(vote.RelatedID, bson.M{"$inc": bson.M{field: value}})
		if err != nil {
			return err
		}

		// Find related comment
		comment, err := comments.FindId(deps.Container, vote.RelatedID)
		if err != nil {
			return err
		}
		userID = comment.UserId
	case "post":
		err = deps.Container.Mgo().C("posts").UpdateId(vote.RelatedID, bson.M{"$inc": bson.M{field: value}})
		if err != nil {
			return err
		}

		post, err := post.FindId(deps.Container, vote.RelatedID)
		if err != nil {
			return err
		}
		userID = post.UserId
	}

	// No gamification for eventual votes from comment's author
	if vote.UserID == userID {
		return nil
	}

	factor := -1
	if vote.Deleted != nil {
		factor = 1
	}
	switch vote.Value {
	case "concise", "useful":
		err = pipeErr(
			gaming.IncreaseUserSwords(deps.Container, vote.UserID, 1*factor),
			gaming.IncreaseUserSwords(deps.Container, userID, 2*factor),
		)
	case "offtopic", "wordy":
		err = pipeErr(
			gaming.IncreaseUserSwords(deps.Container, userID, 1*factor*-1),
		)
	}

	return err
}

func onCommentDelete(e pool.Event) error {
	cid := e.Params["id"].(bson.ObjectId)
	pid := e.Params["post_id"].(bson.ObjectId)

	notify.Transmit <- notify.Socket{
		Chan:   "feed",
		Action: "action",
		Params: map[string]interface{}{
			"fire": "delete-comment",
			"id":   pid.Hex(),
		},
	}

	notify.Transmit <- notify.Socket{
		Chan:   "post",
		Action: pid.Hex(),
		Params: map[string]interface{}{
			"fire": "delete-comment",
			"id":   cid.Hex(),
		},
	}

	if e.Sign != nil {
		audit("comment", cid, "delete", *e.Sign)
	}

	return nil
}

func onPostComment(e pool.Event) error {
	comment, err := comments.FindId(deps.Container, e.Params["id"].(bson.ObjectId))
	if err != nil {
		return err
	}

	if comment.ReplyType == "comment" {
		ref, err := comments.FindId(deps.Container, comment.RelatedID())
		if err != nil {
			return err
		}

		if ref.UserId != comment.UserId {
			notify.Database <- notify.Notification{
				UserId:    ref.UserId,
				Type:      "comment",
				RelatedId: comment.Id,
				Users:     []bson.ObjectId{comment.UserId},
			}
		}
	}

	if comment.ReplyType == "post" {
		post, err := post.FindId(deps.Container, comment.RelatedID())
		if err != nil {
			return err
		}

		if post.UserId != comment.UserId {
			notify.Database <- notify.Notification{
				UserId:    post.UserId,
				Type:      "comment",
				RelatedId: comment.Id,
				Users:     []bson.ObjectId{comment.UserId},
			}
		}

		notify.Transmit <- notify.Socket{
			Chan:   "feed",
			Action: "action",
			Params: map[string]interface{}{
				"fire":    "new-comment",
				"id":      comment.ReplyTo.Hex(),
				"user_id": comment.UserId.Hex(),
			},
		}

		notify.Transmit <- notify.Socket{
			Chan: fmt.Sprintf("post.%s", comment.ReplyTo.Hex()),
			Params: map[string]interface{}{
				"action":     "new-comment",
				"comment_id": comment.Id.Hex(),
			},
		}
	}

	return nil
}

func onCommentUpdate(e pool.Event) error {
	cid := e.Params["id"].(bson.ObjectId)
	pid := e.Params["post_id"].(bson.ObjectId)
	notify.Transmit <- notify.Socket{
		Chan:   "post",
		Action: pid.Hex(),
		Params: map[string]interface{}{
			"fire": "comment-updated",
			"id":   cid.Hex(),
		},
	}

	if e.Sign != nil {
		audit("comment", cid, "update", *e.Sign)
	}

	return nil
}
