package comments

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
)

var CommentNotFound = errors.New("Comment has not been found by given criteria.")

func FindId(deps Deps, id bson.ObjectId) (comment Comment, err error) {
	err = deps.Mgo().C("comments").FindId(id).One(&comment)
	return
}
