package shell

import (
	"github.com/abiosoft/ishell"
	"github.com/fernandez14/spartangeek-blacker/core/events"
	"github.com/fernandez14/spartangeek-blacker/core/post"
	"gopkg.in/mgo.v2/bson"
)

func TestEventHandler(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	c.Println("Testing events mechanism")

	p := post.Post{Id: bson.ObjectIdHex("59b9a86ccdab0b530f68259b")}

	events.In <- events.PostNew(p.Id)
}
