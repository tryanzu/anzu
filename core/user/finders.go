package user

import (
	"errors"
	"github.com/fernandez14/spartangeek-blacker/core/common"
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

func FindEmail(deps Deps, email string) (user User, err error) {
	err = deps.Mgo().C("users").Find(bson.M{"email": email}).One(&user)
	if err != nil {
		return user, UserNotFound
	}

	return
}

func FindList(deps Deps, scopes ...common.Scope) (users Users, err error) {
	err = deps.Mgo().C("users").Find(common.ByScope(scopes...)).All(&users)
	return
}
