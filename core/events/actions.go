package events

import (
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/board/votes"
	"gopkg.in/mgo.v2/bson"
)

func PostNew(id bson.ObjectId) Event {
	return Event{
		Name: POSTS_NEW,
		Params: map[string]interface{}{
			"id": id,
		},
	}
}

func PostView(sign UserSign, id bson.ObjectId) Event {
	return Event{
		Name: POST_VIEW,
		Sign: &sign,
		Params: map[string]interface{}{
			"id": id,
		},
	}
}

func PostsReached(sign UserSign, list []bson.ObjectId) Event {
	return Event{
		Name: POSTS_REACHED,
		Sign: &sign,
		Params: map[string]interface{}{
			"list": list,
		},
	}
}

func PostComment(id bson.ObjectId) Event {
	return Event{
		Name: POSTS_COMMENT,
		Params: map[string]interface{}{
			"id": id,
		},
	}
}

func NewFlag(id bson.ObjectId) Event {
	return Event{
		Name: NEW_FLAG,
		Params: map[string]interface{}{
			"id": id,
		},
	}
}

func DeletePost(sign UserSign, id bson.ObjectId) Event {
	return Event{
		Name: POST_DELETED,
		Sign: &sign,
		Params: map[string]interface{}{
			"id": id,
		},
	}
}

func DeleteComment(sign UserSign, postId, id bson.ObjectId) Event {
	return Event{
		Name: COMMENT_DELETE,
		Sign: &sign,
		Params: map[string]interface{}{
			"id":      id,
			"post_id": postId,
		},
	}
}

func UpdateComment(sign UserSign, postId, id bson.ObjectId) Event {
	return Event{
		Name: COMMENT_UPDATE,
		Sign: &sign,
		Params: map[string]interface{}{
			"id":      id,
			"post_id": postId,
		},
	}
}

func Vote(vote votes.Vote) Event {
	return Event{
		Name: VOTE,
		Params: map[string]interface{}{
			"vote": vote,
		},
	}
}

func RawEmit(channel, event string, params map[string]interface{}) Event {
	return Event{
		Name: RAW_EMIT,
		Params: map[string]interface{}{
			"channel": channel,
			"event":   event,
			"params":  params,
		},
	}
}

func TrackMention(userID, relatedID bson.ObjectId, usersID []bson.ObjectId) Event {
	return Event{
		Name: NEW_MENTION,
		Params: map[string]interface{}{
			"user_id":    userID,
			"related_id": relatedID,
			"users":      usersID,
		},
	}
}

func TrackActivity(m model.Activity) Event {
	return Event{
		Name: RECENT_ACTIVITY,
		Params: map[string]interface{}{
			"activity": m,
		},
	}
}
