package comments

import (
	"html"
	"time"

	"github.com/tryanzu/core/core/content"
	"gopkg.in/mgo.v2/bson"
)

// Delete comment.
func Delete(deps Deps, c Comment) error {
	if c.Deleted != nil {
		return nil
	}
	err := deps.Mgo().C("comments").UpdateId(c.Id, bson.M{
		"$set": bson.M{"deleted_at": time.Now()},
	})
	if err != nil {
		return err
	}
	if c.ReplyType == "post" {
		err = deps.Mgo().C("posts").UpdateId(c.ReplyTo, bson.M{"$inc": bson.M{"comments.count": -1}})
		return err
	}
	return nil
}

func DeletePostComments(deps Deps, postID bson.ObjectId) error {
	_, err := deps.Mgo().C("comments").UpdateAll(
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
	if c.Id.Valid() == false {
		c.Id = bson.NewObjectId()
		c.Created = time.Now()
	}

	if c.ReplyType == "comment" && c.PostId.Valid() == false {
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
	changes, err := deps.Mgo().C("comments").UpsertId(c.Id, bson.M{"$set": c})
	if err != nil {
		return
	}

	if changes.Matched == 0 {
		if c.ReplyType == "post" {
			err = deps.Mgo().C("posts").UpdateId(c.ReplyTo, bson.M{
				"$inc":      bson.M{"comments.count": 1},
				"$set":      bson.M{"updated_at": time.Now()},
				"$addToSet": bson.M{"users": c.UserId},
			})
		} else {
			err = deps.Mgo().C("posts").UpdateId(c.PostId, bson.M{
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
