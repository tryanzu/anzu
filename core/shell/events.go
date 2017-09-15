package shell

import (
	"github.com/abiosoft/ishell"
	"github.com/fernandez14/spartangeek-blacker/core/events"
	_ "github.com/fernandez14/spartangeek-blacker/core/post"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func TestEventHandler(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	c.Println("Testing events mechanism")

	for {
		events.In <- events.PostNew(bson.ObjectIdHex("59b9a86ccdab0b530f68259b"))
		time.Sleep(time.Millisecond * 100)
	}
}
