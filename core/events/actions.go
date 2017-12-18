package events

import (
	"github.com/fernandez14/spartangeek-blacker/model"
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

func TrackActivity(m model.Activity) Event {
	return Event{
		Name: RECENT_ACTIVITY,
		Params: map[string]interface{}{
			"activity": m,
		},
	}
}
