package user

import (
	"errors"
	"github.com/tidwall/buntdb"
	"github.com/tryanzu/core/core/common"
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

func FindNames(deps Deps, list ...bson.ObjectId) (common.UsersStringMap, error) {
	hash := common.UsersStringMap{}
	missing := []bson.ObjectId{}

	// Attempt to fill hashmap using cache layer first.
	deps.BuntDB().View(func(tx *buntdb.Tx) error {
		for _, id := range list {
			v, err := tx.Get("user:" + id.Hex() + ":names")
			if err == nil {
				hash[id] = v
				continue
			}

			// Append to list of missing keys
			missing = append(missing, id)
		}

		return nil
	})

	if len(missing) == 0 {
		return hash, nil
	}

	users, err := FindList(deps, common.WithinID(missing))
	if err != nil {
		return hash, err
	}

	err = deps.BuntDB().Update(users.UpdateBuntCache)
	if err != nil {
		return hash, err
	}

	for _, u := range users {
		hash[u.Id] = u.UserName
	}

	return hash, nil
}
