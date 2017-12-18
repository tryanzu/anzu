package events

import (
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
			err = deps.Container.Mgo().C("activity").Insert(activity)
			return
		},
	}
}
