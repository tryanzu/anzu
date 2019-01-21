package events

import (
	notify "github.com/tryanzu/core/board/notifications"
	pool "github.com/tryanzu/core/core/events"
)

// Bind event handlers for global unrelated actions...
func globalEvents() {
	pool.On <- pool.EventHandler{
		On: pool.RAW_EMIT,
		Handler: func(e pool.Event) (err error) {

			// Just broadcast it to transmit channel
			m := notify.Socket{
				Chan:   e.Params["channel"].(string),
				Action: e.Params["event"].(string),
				Params: e.Params["params"].(map[string]interface{}),
			}
			notify.Transmit <- m
			return
		},
	}
}
