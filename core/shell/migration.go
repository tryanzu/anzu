package shell

import (
	"context"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/tryanzu/core/board/comments"
	"github.com/tryanzu/core/deps"
	"go.mongodb.org/mongo-driver/bson"
)

func MigrateComments(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db := deps.Container.Mgo()
	collection := db.Collection("comments")

	filter := bson.M{"reply_type": bson.M{"$exists": false}}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.Printf("Error finding comments: %v\n", err)
		return
	}
	defer cursor.Close(ctx)

	var comment comments.Comment
	c.ProgressBar().Indeterminate(true)
	c.ProgressBar().Start()

	for cursor.Next(ctx) {
		err := cursor.Decode(&comment)
		if err != nil {
			c.Printf("Error decoding comment: %v\n", err)
			continue
		}

		updateCtx, updateCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = collection.UpdateOne(updateCtx, bson.M{"_id": comment.Id}, bson.M{"$set": bson.M{
			"reply_type": "post",
			"reply_to":   comment.PostId,
		}})
		updateCancel()

		if err != nil {
			c.Println("Could not migrate comment", err)
		}
	}
	c.ProgressBar().Stop()
}
