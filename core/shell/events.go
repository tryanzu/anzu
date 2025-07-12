package shell

import (
	"github.com/abiosoft/ishell"
	_ "github.com/tryanzu/core/board/comments"
	_ "github.com/tryanzu/core/board/posts"
	"github.com/tryanzu/core/core/events"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestEventHandler(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	c.Println("Testing events mechanism")

	objID, _ := primitive.ObjectIDFromHex("59a9a33bcdab0b5dcb31d4b0")
	events.In <- events.PostComment(objID)
}
