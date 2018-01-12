package shell

import (
	_ "github.com/tryanzu/core/board/comments"
	_ "github.com/tryanzu/core/board/posts"
	"github.com/tryanzu/core/core/events"
	"gopkg.in/abiosoft/ishell.v2"
	"gopkg.in/mgo.v2/bson"
)

func TestEventHandler(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	c.Println("Testing events mechanism")

	events.In <- events.PostComment(bson.ObjectIdHex("59a9a33bcdab0b5dcb31d4b0"))
}
