package events

import (
	"log"

	"gopkg.in/mgo.v2/bson"
)

type Handler func(Event) error

// Input channel for incoming events.
var In chan Event

// On "event" channel. Register event handlers using channels.
var On chan EventHandler

// Map of handlers that will react to events.
var Handlers map[string][]Handler

type EventHandler struct {
	On      string
	Handler Handler
}

type Event struct {
	Name   string
	Sign   *UserSign
	Params map[string]interface{}
}

type UserSign struct {
	Reason string
	UserID bson.ObjectId
}

func execHandlers(list []Handler, event Event) {
	var err error
	for h := range list {
		err = list[h](event)
		if err != nil {
			panic(err)
		}
	}
}

func sink(in chan Event, on chan EventHandler) {
	for {
		select {
		case event := <-in: // For incoming events spawn a goroutine running handlers.
			log.Printf("Incoming event: %+v", event)
			if ls, exists := Handlers[event.Name]; exists {
				go execHandlers(ls, event)
			}
		case h := <-on: // Register new handlers.
			if _, exists := Handlers[h.On]; !exists {
				Handlers[h.On] = []Handler{}
			}

			Handlers[h.On] = append(Handlers[h.On], h.Handler)
		}
	}
}

// init channel for input events, consumers & map of handlers.
func init() {
	In = make(chan Event, 10)
	On = make(chan EventHandler)
	Handlers = make(map[string][]Handler)

	go sink(In, On)
}
