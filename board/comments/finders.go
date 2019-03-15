package comments

import (
	"errors"

	"github.com/tryanzu/core/core/common"
	"github.com/tryanzu/core/core/content"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var CommentNotFound = errors.New("Comment has not been found by given criteria.")

func FetchCount(d Deps, query common.Query) (c int, err error) {
	c, err = query(d.Mgo().C("comments")).Limit(0).Count()
	return
}

func FetchBy(deps Deps, query common.Query) (CommentsSet, error) {
	c, err := query(deps.Mgo().C("comments")).Limit(0).Count()
	if err != nil {
		return CommentsSet{}, err
	}

	var list Comments
	err = query(deps.Mgo().C("comments")).All(&list)
	if err != nil {
		return CommentsSet{}, err
	}

	var processed content.Parseable
	for n, c := range list {
		processed, err = content.Postprocess(deps, c)
		if err != nil {
			return CommentsSet{}, err
		}

		list[n] = processed.(Comment)
	}

	return CommentsSet{
		List:  list,
		Count: c,
	}, nil
}

func Post(id bson.ObjectId, limit, offset int, reverse bool, before *bson.ObjectId, after *bson.ObjectId) common.Query {
	return func(col *mgo.Collection) *mgo.Query {
		criteria := bson.M{
			"reply_type": "post",
			"reply_to":   id,
			"deleted_at": bson.M{"$exists": false},
		}

		if before != nil {
			criteria["_id"] = bson.M{"$lt": before}
			offset = 0
		}

		if after != nil {
			criteria["_id"] = bson.M{"$gt": after}
			offset = 0
		}

		return col.Find(criteria).Limit(limit).Skip(offset).Sort("-created_at")
		// .Sort("-votes.up", "votes.down", "-created_at")
	}
}

func User(id bson.ObjectId, limit, offset int) common.Query {
	return func(col *mgo.Collection) *mgo.Query {
		return col.Find(bson.M{"user_id": id}).Limit(limit).Skip(offset).Sort("-created_at")
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
		{"$match": bson.M{"reply_type": "comment", "reply_to": bson.M{"$in": list.IDList()}}},
		{"$sort": bson.M{"-created_at": 1}},
		{"$group": bson.M{"_id": "$reply_to", "count": bson.M{"$sum": 1}, "list": bson.M{"$push": "$$ROOT"}}},
		{"$project": bson.M{"count": 1, "list": bson.M{"$slice": []interface{}{"$list", 0, max}}}},
	}).All(&lists)
	return
}
