package user

import (
	"gopkg.in/mgo.v2/bson"
)

func ResetNotifications(d deps, id bson.ObjectId) (err error) {
	err = d.Mgo().C("users").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"notifications": 0}})
	return
}
