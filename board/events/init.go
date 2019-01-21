package events

import pool "github.com/tryanzu/core/core/events"

func init() {
	globalEvents()
	activityEvents()
	commentsEvents()
	postsEvents()
	mentionEvents()
}

func register(list []pool.EventHandler) {
	for _, h := range list {
		pool.On <- h
	}
}

func pipeErr(err ...error) error {
	for _, e := range err {
		if e != nil {
			return e
		}
	}
	return nil
}
