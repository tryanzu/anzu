package events

import (
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

	// Find related comment
	comment, err := comments.FindId(deps.Container, vote.RelatedID)
	if err != nil {
		return err
	}

	// Increment value by given state.
	value := 1
	if vote.Deleted != nil {
		value = -1
	}

	err = deps.Container.Mgo().C("comments").UpdateId(vote.RelatedID, bson.M{"$inc": bson.M{field: value}})
	if err != nil {
		return err
	}

	if vote.UserID != comment.UserId {
		if vote.Deleted == nil {
			err = gaming.UserHasVoted(deps.Container, vote.UserID)
			if err != nil {
				return err
			}

			err = gaming.UserReceivedVote(deps.Container, comment.UserId)
			if err != nil {
				return err
			}
		} else {
			err = gaming.UserRemovedVote(deps.Container, vote.UserID)
			if err != nil {
				return err
			}

			err = gaming.UserRevokedVote(deps.Container, comment.UserId)
			if err != nil {
				return err
			}
		}
	}

	return nil
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

	notify.Transmit <- notify.Socket{"feed", "action", map[string]interface{}{
		"fire":    "new-comment",
		"id":      comment.ReplyTo.Hex(),
		"user_id": comment.UserId.Hex(),
	}}

	return gaming.UserHasCommented(deps.Container, comment.UserId)
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
