package shell

import (
	"github.com/abiosoft/ishell"
	_ "github.com/fernandez14/spartangeek-blacker/board/comments"
	"github.com/fernandez14/spartangeek-blacker/core/events"
	_ "github.com/fernandez14/spartangeek-blacker/core/post"
	"gopkg.in/mgo.v2/bson"
)

func TestEventHandler(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	c.Println("Testing events mechanism")

	events.In <- events.PostComment(bson.ObjectIdHex("59a9a33bcdab0b5dcb31d4b0"))
}
