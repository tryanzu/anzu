package events

import (
	notify "github.com/fernandez14/spartangeek-blacker/board/notifications"
	pool "github.com/fernandez14/spartangeek-blacker/core/events"
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
