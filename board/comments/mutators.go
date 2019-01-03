package comments

import (
	"html"
	"time"

	"github.com/tryanzu/core/core/content"
	"gopkg.in/mgo.v2/bson"
)

// UpsertComment performs validations before upserting data struct
func UpsertComment(deps Deps, c Comment) (comment Comment, err error) {
	if c.Id.Valid() == false {
		c.Id = bson.NewObjectId()
		c.Created = time.Now()
	}

	if c.ReplyType == "comment" {
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
	_, err = deps.Mgo().C("comments").UpsertId(c.Id, bson.M{"$set": c})
	if err != nil {
		return
	}

	if c.ReplyType == "post" {
		err = deps.Mgo().C("posts").UpdateId(c.ReplyTo, bson.M{
			"$inc":      bson.M{"comments.count": 1},
			"$addToSet": bson.M{"users": c.UserId},
		})
	} else {
		err = deps.Mgo().C("posts").UpdateId(c.PostId, bson.M{
			"$addToSet": bson.M{"users": c.UserId},
		})
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
