package comments

import (
	"errors"
	"github.com/fernandez14/spartangeek-blacker/core/common"
	"gopkg.in/mgo.v2/bson"
)

var CommentNotFound = errors.New("Comment has not been found by given criteria.")

func FindId(deps Deps, id bson.ObjectId) (comment Comment, err error) {
	err = deps.Mgo().C("comments").FindId(id).One(&comment)
	return
}

func FindList(deps Deps, scopes ...common.Scope) (list Comments, err error) {
	err = deps.Mgo().C("comments").Find(common.ByScope(scopes...)).All(&list)
	return
}
