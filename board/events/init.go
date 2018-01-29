package events

import pool "github.com/tryanzu/core/core/events"

func init() {
	globalEvents()
	activityEvents()
	commentsEvents()
	postsEvents()
}

func register(list []pool.EventHandler) {
	for _, h := range list {
		pool.On <- h
	}
}
