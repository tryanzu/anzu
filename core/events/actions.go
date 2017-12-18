package events

import (
	"github.com/fernandez14/spartangeek-blacker/board/legacy/model"
	notify "github.com/fernandez14/spartangeek-blacker/board/notifications"
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

func UpvoteCommentRemove(id bson.ObjectId) Event {
	return Event{
		Name: COMMENT_UPVOTE_REMOVE,
		Params: map[string]interface{}{
			"id": id,
		},
	}
}

func UpvoteComment(id bson.ObjectId) Event {
	return Event{
		Name: COMMENT_UPVOTE,
		Params: map[string]interface{}{
			"id": id,
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
