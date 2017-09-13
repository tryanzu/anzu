package post

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
)

var PostNotFound = errors.New("Post has not been found by given criteria.")

func FindId(deps Deps, id bson.ObjectId) (post Post, err error) {
	err = deps.Mgo().C("posts").FindId(id).One(&post)
	return
}
