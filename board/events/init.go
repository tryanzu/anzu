package events

import (
	"github.com/op/go-logging"
	pool "github.com/tryanzu/core/core/events"
)

var (
	log = logging.MustGetLogger("main")
)

func init() {
	globalEvents()
	activityEvents()
	commentsEvents()
	postsEvents()
	mentionEvents()
	flagHandlers()
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
