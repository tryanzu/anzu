package events

import (
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
