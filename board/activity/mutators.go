package activity

import (
	"time"
)

// Track activity.
func Track(d deps, activity M) (err error) {
	activity.Created = time.Now()
	err = d.Mgo().C("activity").Insert(&activity)
	return
}
