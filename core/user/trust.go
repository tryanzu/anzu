package user

import (
	"gopkg.in/mgo.v2/bson"
)

func CanBeTrusted(user User) bool {
	return user.Warnings < 6
}

func IsBanned(d deps, id bson.ObjectId) bool {
	ledis := d.LedisDB()
	k := []byte("ban:")
	k = append(k, []byte(id)...)
	n, err := ledis.Exists(k)
	if err != nil {
		panic(err)
	}
	return n == 1
}