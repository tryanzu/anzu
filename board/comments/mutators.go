package comments

import (
	"context"
	"html"
	"time"

	"github.com/tryanzu/core/core/content"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Delete comment.
func Delete(deps Deps, c Comment) error {
	if c.Deleted != nil {
		return nil
	}
	ctx := context.TODO()
	_, err := deps.Mgo().Collection("comments").UpdateOne(ctx, bson.M{"_id": c.Id}, bson.M{
		"$set": bson.M{"deleted_at": time.Now()},
	})
	if err != nil {
		return err
	}
	if c.ReplyType == "post" {
		_, err = deps.Mgo().Collection("posts").UpdateOne(ctx, bson.M{"_id": c.ReplyTo}, bson.M{"$inc": bson.M{"comments.count": -1}})
		return err
	}
	return nil
}

func DeletePostComments(deps Deps, postID primitive.ObjectID) error {
	ctx := context.TODO()
	_, err := deps.Mgo().Collection("comments").UpdateMany(ctx,
		bson.M{"$or": []bson.M{
			{"post_id": postID},
			{"reply_to": postID},
		}}, bson.M{
			"$set": bson.M{"deleted_at": time.Now()},
		})
	return err
}

// UpsertComment performs validations before upserting data struct
func UpsertComment(deps Deps, c Comment) (comment Comment, err error) {
	ctx := context.TODO()
	isNew := false
	if c.Id.IsZero() {
		c.Id = primitive.NewObjectID()
		c.Created = time.Now()
		isNew = true
	}

	if c.ReplyType == "comment" && c.PostId.IsZero() {
		id := c.ReplyTo
		for {
			var ref Comment
			ref, err = FindId(deps, id)
			if err != nil {
				return
			}
			id = ref.ReplyTo
			if ref.ReplyType == "post" {
				c.PostId = ref.ReplyTo
				break
			}
			continue
		}
	}

	c.Content = html.EscapeString(c.Content)
	c.Updated = time.Now()

	// Pre-process comment content.
	processed, err := content.Preprocess(deps, c)
	if err != nil {
		return
	}

	c = processed.(Comment)
	upsertTrue := true
	_, err = deps.Mgo().Collection("comments").ReplaceOne(ctx, bson.M{"_id": c.Id}, c, &options.ReplaceOptions{Upsert: &upsertTrue})
	if err != nil {
		return
	}

	if isNew {
		if c.ReplyType == "post" {
			_, err = deps.Mgo().Collection("posts").UpdateOne(ctx, bson.M{"_id": c.ReplyTo}, bson.M{
				"$inc":      bson.M{"comments.count": 1},
				"$set":      bson.M{"updated_at": time.Now()},
				"$addToSet": bson.M{"users": c.UserId},
			})
		} else {
			_, err = deps.Mgo().Collection("posts").UpdateOne(ctx, bson.M{"_id": c.PostId}, bson.M{
				"$addToSet": bson.M{"users": c.UserId},
				"$set":      bson.M{"updated_at": time.Now()},
			})
		}
	}

	if err != nil {
		return
	}

	// Pre-process comment content.
	processed, err = content.Postprocess(deps, c)
	if err != nil {
		return
	}
	c = processed.(Comment)
	comment = c
	return
}
