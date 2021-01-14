package events

import (
	"errors"
	"fmt"
	"time"

	"github.com/tryanzu/core/board/comments"
	"github.com/tryanzu/core/board/notifications"
	notify "github.com/tryanzu/core/board/notifications"
	post "github.com/tryanzu/core/board/posts"
	"github.com/tryanzu/core/board/votes"
	"github.com/tryanzu/core/core/config"
	pool "github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/mail"
	"github.com/tryanzu/core/core/user"
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
	factor := 1
	if vote.Deleted != nil {
		factor = -1
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
		factor = factor * 4
	case "post":
		err = deps.Container.Mgo().C("posts").UpdateId(vote.RelatedID, bson.M{"$inc": bson.M{field: value}})
		if err != nil {
			return err
		}
		p, err := post.FindId(deps.Container, vote.RelatedID)
		if err != nil {
			return err
		}
		userID = p.UserId
		factor = factor * 2
	}

	// No gamification for eventual votes from comment's author
	if vote.UserID == userID {
		return nil
	}
	rules := config.C.Rules()
	if rule, exists := rules.Reactions[vote.Value]; exists {
		rewards, err := rule.Rewards()
		if err != nil {
			return err
		}
		if rewards.Provider > 0 {
			err = gaming.IncreaseUserSwords(deps.Container, vote.UserID, int(rewards.Provider))
		}

		if rewards.Receiver > 0 {
			err = gaming.IncreaseUserSwords(deps.Container, userID, int(rewards.Receiver))
		}
	}

	switch vote.Value {
	case "concise", "useful":
		err = pipeErr(
			//gaming.IncreaseUserSwords(deps.Container, vote.UserID, 1*factor),
			gaming.IncreaseUserSwords(deps.Container, userID, 2*factor),
		)
	case "offtopic", "wordy":
		err = pipeErr(
			gaming.IncreaseUserSwords(deps.Container, userID, 1*factor/2*-1),
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
	switch comment.ReplyType {
	case "comment":
		ref, err := comments.FindId(deps.Container, comment.RelatedID())
		if err != nil || ref.UserId == comment.UserId {
			if ref.UserId == comment.UserId {
				log.Debugf("reply to a comment of self	id=%s", comment.RelatedID().Hex())
			} else {
				log.Errorf("comments.FindId 	err=%s", err)
			}
			return err
		}
		notify.Database <- notify.Notification{
			UserId:    ref.UserId,
			Type:      "comment",
			RelatedId: comment.Id,
			Users:     []bson.ObjectId{comment.UserId},
		}
		p, err := post.FindId(deps.Container, ref.RelatedPost())
		if err != nil {
			log.Errorf("post.FindId 	err=%s", err)
			return err
		}
		usr, err := user.FindId(deps.Container, ref.UserId)
		if err != nil {
			log.Errorf("user.FindId 	err=%s", err)
			return err
		}
		seen := usr.Seen
		if usr.EmailNotifications && (seen == nil || seen.Add(time.Minute*15).Before(time.Now())) {
			m, err := notifications.SomeoneCommentedYourCommentEmail(p, ref, usr)
			if err != nil {
				return err
			}

			// Send email message.
			mail.In <- m
		}
		return nil
	case "post":
		p, err := post.FindId(deps.Container, comment.RelatedID())
		if err != nil {
			return err
		}
		if p.UserId != comment.UserId {
			usr, err := user.FindId(deps.Container, p.UserId)
			if err != nil {
				return err
			}
			seen := usr.Seen
			if usr.EmailNotifications && (seen == nil || seen.Add(time.Minute*15).Before(time.Now())) {
				m, err := notifications.SomeoneCommentedYourPostEmail(p, usr)
				if err != nil {
					return err
				}

				// Send email message.
				mail.In <- m
			}
			notify.Database <- notify.Notification{
				UserId:    p.UserId,
				Type:      "comment",
				RelatedId: comment.Id,
				Users:     []bson.ObjectId{comment.UserId},
			}
		}

		c, err := comments.FetchCount(deps.Container, comments.Post(p.Id, 0, 0, false, nil, nil))
		if err != nil {
			return err
		}

		notify.Transmit <- notify.Socket{
			Chan:   "feed",
			Action: "action",
			Params: map[string]interface{}{
				"fire":    "new-comment",
				"count":   c,
				"id":      comment.ReplyTo.Hex(),
				"user_id": comment.UserId.Hex(),
			},
		}

		notify.Transmit <- notify.Socket{
			Chan: fmt.Sprintf("post.%s", comment.ReplyTo.Hex()),
			Params: map[string]interface{}{
				"action":     "new-comment",
				"count":      c,
				"comment_id": comment.Id.Hex(),
			},
		}
		return nil
	}
	return errors.New("invalid onPostComment comment.ReplyType " + comment.ReplyType)
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
