package events

import (
	ev "github.com/tryanzu/core/core/events"
)

// Bind event handlers for flag related actions...
func flagHandlers() {
	ev.On <- ev.EventHandler{
		On: ev.NEW_FLAG,
		Handler: func(e ev.Event) error {
			// WIP
			return nil
		},
	}
}
