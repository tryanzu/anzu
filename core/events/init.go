package events

type Handler func(Event) error

// The input channel will receive
var In chan Event

// Map of handlers that will react to events.
var Handlers map[string][]Handler

type Event struct {
	Name   string
	Params map[string]interface{}
}

func On(event string, h Handler) {
	if _, exists := Handlers[event]; !exists {
		Handlers[event] = []Handler{}
	}

	Handlers[event] = append(Handlers[event], h)
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

func inputEvents(ch chan Event) {
	for event := range ch {
		if ls, exists := Handlers[event.Name]; exists {
			go execHandlers(ls, event)
		}
	}
}

// Initialize channel of input events, consumer & map of handlers.
func init() {
	In = make(chan Event, 10)
	Handlers = make(map[string][]Handler)

	go inputEvents(In)
}
