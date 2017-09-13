package shell

import (
	"github.com/abiosoft/ishell"
	"github.com/fernandez14/spartangeek-blacker/core/events"
	"math/rand"
	"time"
)

func TestEventHandler(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	c.Println("Testing events mechanism")

	// Define listeners
	events.On("post:new", func(e events.Event) error {
		c.Println("[new-post]: rand id is", e.Params["n"])
		return nil
	})

	n := 0
	for {
		source := rand.NewSource(time.Now().UnixNano())
		r := rand.New(source)

		events.In <- events.Event{
			Name: "new-post",
			Params: map[string]interface{}{
				"id":   n,
				"rand": r.Intn(10),
			},
		}

		n = n + 1
	}
}
