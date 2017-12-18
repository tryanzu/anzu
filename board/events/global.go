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
			m := e.Params["socket"].(notify.Socket)
			notify.Transmit <- m
			return
		},
	}
}
