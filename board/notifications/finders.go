package notifications

import (
	"errors"

	"github.com/tryanzu/core/core/common"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var NotificationNotFound = errors.New("Notification has not been found by given criteria.")

func FindId(deps Deps, id bson.ObjectId) (notification Notification, err error) {
	err = deps.Mgo().C("notifications").FindId(id).One(&notification)
	return
}

// Fetch multiple leads by conditions
func FetchBy(deps Deps, query common.Query) (list Notifications, err error) {
	err = query(deps.Mgo().C("notifications")).All(&list)
	return
}

func UserID(id bson.ObjectId, take, skip int) common.Query {
	return func(col *mgo.Collection) *mgo.Query {
		return col.Find(bson.M{"user_id": id}).Limit(take).Skip(skip).Sort("-updated_at")
	}
}
