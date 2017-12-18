package post

import (
	"errors"
	"github.com/tryanzu/core/core/common"
	"gopkg.in/mgo.v2/bson"
)

var PostNotFound = errors.New("Post has not been found by given criteria.")

func FindId(deps Deps, id bson.ObjectId) (post Post, err error) {
	err = deps.Mgo().C("posts").FindId(id).One(&post)
	return
}

func FindList(deps Deps, scopes ...common.Scope) (list Posts, err error) {
	err = deps.Mgo().C("posts").Find(common.ByScope(scopes...)).All(&list)
	return
}
