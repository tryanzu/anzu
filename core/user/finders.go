package user

import (
	"errors"

	"github.com/tryanzu/core/core/common"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var UserNotFound = errors.New("User has not been found by given criteria.")

func FindId(d deps, id bson.ObjectId) (user User, err error) {
	err = d.Mgo().C("users").FindId(id).One(&user)
	if err != nil {
		return user, UserNotFound
	}

	return
}

func FindEmail(d deps, email string) (user User, err error) {
	err = d.Mgo().C("users").Find(bson.M{"email": email}).One(&user)
	if err != nil {
		return user, UserNotFound
	}

	return
}

func FindList(d deps, scopes ...common.Scope) (users Users, err error) {
	err = d.Mgo().C("users").Find(common.ByScope(scopes...)).All(&users)
	return
}

func FetchBy(d deps, query common.Query) (UsersSet, error) {
	c, err := query(d.Mgo().C("users")).Limit(0).Count()
	if err != nil {
		return UsersSet{}, err
	}

	list := Users{}
	err = query(d.Mgo().C("users")).All(&list)
	if err != nil {
		return UsersSet{}, err
	}
	return UsersSet{
		Count: c,
		List:  list,
	}, nil
}

func Page(limit int, reverse bool, before *bson.ObjectId, after *bson.ObjectId) common.Query {
	return func(col *mgo.Collection) *mgo.Query {
		criteria := bson.M{
			"deleted_at": bson.M{"$exists": false},
		}

		if before != nil {
			criteria["_id"] = bson.M{"$lt": before}
		}

		if after != nil {
			criteria["_id"] = bson.M{"$gt": after}
		}

		return col.Find(criteria).Limit(limit).Skip(0).Sort("-created_at")
	}
}

func FindNames(d deps, list ...bson.ObjectId) (common.UsersStringMap, error) {
	hash := common.UsersStringMap{}
	missing := []bson.ObjectId{}

	for _, id := range list {
		v, err := d.LedisDB().Get([]byte("user:" + id.Hex() + ":names"))
		if err == nil && len(v) > 0 {
			hash[id] = string(v)
			continue
		}

		// Append to list of missing keys
		missing = append(missing, id)
	}

	if len(missing) == 0 {
		return hash, nil
	}

	users, err := FindList(d, common.WithinID(missing))
	if err != nil {
		return hash, err
	}

	err = users.UpdateCache(d)
	if err != nil {
		return hash, err
	}

	for _, u := range users {
		hash[u.Id] = u.UserName
	}

	// Unknown users should be cached like so...
	if len(missing) != len(users) {
		for _, id := range missing {
			if _, exists := hash[id]; exists == false {
				hash[id] = "Unknown"
				err = d.LedisDB().Set([]byte("user:"+id.Hex()+":names"), []byte("Unknown"))
				if err != nil {
					return hash, err
				}
			}
		}
	}

	return hash, nil
}
