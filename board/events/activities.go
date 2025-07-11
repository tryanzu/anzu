package events

import (
	"context"
	"time"

	"github.com/tryanzu/core/board/legacy/model"
	pool "github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/deps"
)

// Bind event handlers for activity related actions...
func activityEvents() {
	pool.On <- pool.EventHandler{
		On: pool.RECENT_ACTIVITY,
		Handler: func(e pool.Event) (err error) {
			activity := e.Params["activity"].(model.Activity)
			activity.Created = time.Now()

			// Attempt to record recent activity.
			_, err = deps.Container.Mgo().Collection("activity").InsertOne(context.Background(), activity)
			return
		},
	}
}
