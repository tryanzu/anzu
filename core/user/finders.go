package user

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
)

var UserNotFound = errors.New("User has not been found by given criteria.")

func FindId(deps Deps, id bson.ObjectId) (user User, err error) {
	err = deps.Mgo().C("users").FindId(id).One(&user)
	if err != nil {
		return user, UserNotFound
	}

	return
}
