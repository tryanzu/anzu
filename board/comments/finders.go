package comments

import (
	"errors"

	"github.com/fernandez14/spartangeek-blacker/core/common"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var CommentNotFound = errors.New("Comment has not been found by given criteria.")

func FetchBy(deps Deps, query common.Query) (list Comments, err error) {
	err = query(deps.Mgo().C("comments")).All(&list)
	return
}

func Post(id bson.ObjectId, limit, offset int) common.Query {
	return func(col *mgo.Collection) *mgo.Query {
		return col.Find(bson.M{"post_id": id, "reply_to": bson.M{"$exists": false}}).Limit(limit).Skip(offset).Sort("-votes.up", "votes.down")
	}
}

func User(id bson.ObjectId, limit, offset int) common.Query {
	return func(col *mgo.Collection) *mgo.Query {
		return col.Find(bson.M{"user_id": id}).Limit(limit).Skip(offset)
	}
}

func FindId(deps Deps, id bson.ObjectId) (comment Comment, err error) {
	err = deps.Mgo().C("comments").FindId(id).One(&comment)
	return
}

func FindList(deps Deps, scopes ...common.Scope) (list Comments, err error) {
	err = deps.Mgo().C("comments").Find(common.ByScope(scopes...)).All(&list)
	return
}

func FindReplies(deps Deps, list Comments, max int) (lists []Replies, err error) {
	err = deps.Mgo().C("comments").Pipe([]bson.M{
		{"$match": bson.M{"reply_to": bson.M{"$in": list.IDList()}}},
		{"$sort": bson.M{"votes.up": 1, "votes.down": -1}},
		{"$group": bson.M{"_id": "$reply_to", "count": bson.M{"$sum": 1}, "list": bson.M{"$push": "$$ROOT"}}},
		//{"$project": bson.M{"count": 1, "list": []interface{}{"$list", 0, max}}},
	}).All(&lists)
	return
}
