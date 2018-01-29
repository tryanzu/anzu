package events

import (
	"github.com/tryanzu/core/board/legacy/model"
	notify "github.com/tryanzu/core/board/notifications"
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

func PostComment(id bson.ObjectId) Event {
	return Event{
		Name: POSTS_COMMENT,
		Params: map[string]interface{}{
			"id": id,
		},
	}
}

func DeleteComment(postId, id bson.ObjectId) Event {
	return Event{
		Name: COMMENT_DELETE,
		Params: map[string]interface{}{
			"id":      id,
			"post_id": postId,
		},
	}
}

func UpdateComment(postId, id bson.ObjectId) Event {
	return Event{
		Name: COMMENT_UPDATE,
		Params: map[string]interface{}{
			"id":      id,
			"post_id": postId,
		},
	}
}

func VoteComment(vote votes.Vote) Event {
	return Event{
		Name: COMMENT_VOTE,
		Params: map[string]interface{}{
			"vote": vote,
		},
	}
}

func RawEmit(channel, event string, params map[string]interface{}) Event {
	return Event{
		Name: RAW_EMIT,
		Params: map[string]interface{}{
			"socket": notify.Socket{channel, event, params},
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
