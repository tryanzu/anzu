package activity

import (
	"context"
	"time"
)

// Track activity.
func Track(d deps, activity M) (err error) {
	activity.Created = time.Now()
	ctx := context.TODO()
	_, err = d.Mgo().Collection("activity").InsertOne(ctx, activity)
	return
}
