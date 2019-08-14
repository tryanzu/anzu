package shell

import (
	"github.com/abiosoft/ishell"
	"github.com/tryanzu/core/board/comments"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

func MigrateComments(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	db := deps.Container.Mgo()
	migratable := db.C("comments").Find(bson.M{"reply_type": bson.M{"$exists": false}}).Sort("-$natural").Iter()
	var comment comments.Comment
	c.ProgressBar().Indeterminate(true)
	c.ProgressBar().Start()
	for migratable.Next(&comment) {
		err := db.C("comments").UpdateId(comment.Id, bson.M{"$set": bson.M{
			"reply_type": "post",
			"reply_to":   comment.PostId,
		}})
		if err != nil {
			c.Println("Could not migrate comment", err)
		}
	}
	c.ProgressBar().Stop()
}
