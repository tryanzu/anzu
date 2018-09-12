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
			On:      pool.COMMENT_VOTE,
			Handler: onCommentVote,
		},
	}

	register(handlers)
}

func onCommentVote(e pool.Event) error {
	vote := e.Params["vote"].(votes.Vote)
	field := vote.DbField()
	d := deps.Container

	// Find related comment
	comment, err := comments.FindId(d, vote.RelatedID)
	if err != nil {
		return err
	}

	// Increment value by given state.
	value := 1
	if vote.Deleted != nil {
		value = -1
	}

	err = d.Mgo().C("comments").UpdateId(vote.RelatedID, bson.M{"$inc": bson.M{field: value}})
	if err != nil {
		return err
	}

	// No gamification for eventual votes from comment's author
	if vote.UserID == comment.UserId {
		return nil
	}

	factor := -1
	if vote.Deleted != nil {
		factor = 1
	}
	switch vote.Direction() {
	case votes.DOWN:
		err = pipeErr(
			gaming.IncreaseUserSwords(d, vote.UserID, 1*factor),
			gaming.IncreaseUserSwords(d, comment.UserId, 2*factor),
		)
	case votes.UP:
		err = pipeErr(
			gaming.IncreaseUserSwords(d, comment.UserId, 5*factor*-1),
		)
	}

	return err
}

func onCommentDelete(e pool.Event) error {
	cid := e.Params["id"].(bson.ObjectId)
	pid := e.Params["post_id"].(bson.ObjectId)

	notify.Transmit <- notify.Socket{"feed", "action", map[string]interface{}{
		"fire": "delete-comment",
		"id":   pid.Hex(),
	}}

	notify.Transmit <- notify.Socket{"post", pid.Hex(), map[string]interface{}{
		"fire": "delete-comment",
		"id":   cid.Hex(),
	}}

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

	notify.Transmit <- notify.Socket{"post", pid.Hex(), map[string]interface{}{
		"fire": "comment-updated",
		"id":   cid.Hex(),
	}}

	return nil
}
