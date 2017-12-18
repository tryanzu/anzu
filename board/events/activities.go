package events

import (
	"time"

	pool "github.com/fernandez14/spartangeek-blacker/core/events"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/model"
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
