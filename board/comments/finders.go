package comments

import (
	"context"
	"errors"

	"github.com/tryanzu/core/board/votes"
	"github.com/tryanzu/core/core/common"
	"github.com/tryanzu/core/core/content"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var CommentNotFound = errors.New("Comment has not been found by given criteria.")

func FetchCount(d Deps, query common.Query) (c int, err error) {
	ctx := context.TODO()
	cursor, err := query(d.Mgo().Collection("comments"), ctx)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)
	count := 0
	for cursor.Next(ctx) {
		count++
	}
	return count, nil
}

func FetchBy(deps Deps, query common.Query) (CommentsSet, error) {
	ctx := context.TODO()
	cursor, err := query(deps.Mgo().Collection("comments"), ctx)
	if err != nil {
		return CommentsSet{}, err
	}
	defer cursor.Close(ctx)

	var list Comments
	err = cursor.All(ctx, &list)
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
		if list[n].Votes == nil {
			list[n].Votes = votes.Votes{}
		}
	}

	return CommentsSet{
		List:  list,
		Count: len(list),
	}, nil
}

func Post(id primitive.ObjectID, limit, offset int, reverse bool, before *primitive.ObjectID, after *primitive.ObjectID) common.Query {
	return func(col *mongo.Collection, ctx context.Context) (*mongo.Cursor, error) {
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

		opts := &options.FindOptions{}
		if limit > 0 {
			opts.SetLimit(int64(limit))
		}
		if offset > 0 {
			opts.SetSkip(int64(offset))
		}
		opts.SetSort(bson.M{"created_at": -1})

		return col.Find(ctx, criteria, opts)
	}
}

func User(id primitive.ObjectID, limit, offset int) common.Query {
	return func(col *mongo.Collection, ctx context.Context) (*mongo.Cursor, error) {
		criteria := bson.M{
			"user_id":    id,
			"deleted_at": bson.M{"$exists": false},
		}

		opts := &options.FindOptions{}
		if limit > 0 {
			opts.SetLimit(int64(limit))
		}
		if offset > 0 {
			opts.SetSkip(int64(offset))
		}
		opts.SetSort(bson.M{"created_at": -1})

		return col.Find(ctx, criteria, opts)
	}
}

func FindId(deps Deps, id primitive.ObjectID) (comment Comment, err error) {
	ctx := context.TODO()
	err = deps.Mgo().Collection("comments").FindOne(ctx, bson.M{"_id": id}).Decode(&comment)
	if err == mongo.ErrNoDocuments {
		err = CommentNotFound
	}
	return
}

func FindList(deps Deps, scopes ...common.Scope) (list Comments, err error) {
	ctx := context.TODO()
	cursor, err := deps.Mgo().Collection("comments").Find(ctx, common.ByScope(scopes...))
	if err != nil {
		return
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &list)
	return
}

func FindReplies(deps Deps, list Comments, max int) (lists []Replies, err error) {
	if len(list) == 0 {
		return []Replies{}, nil
	}

	ctx := context.TODO()
	pipeline := []bson.M{
		{"$match": bson.M{
			"reply_type": "comment",
			"reply_to":   bson.M{"$in": list.IDList()},
			"deleted_at": bson.M{"$exists": false},
		}},
		{"$sort": bson.M{"-created_at": 1}},
		{"$group": bson.M{"_id": "$reply_to", "count": bson.M{"$sum": 1}, "list": bson.M{"$push": "$$ROOT"}}},
		{"$project": bson.M{"count": 1, "list": bson.M{"$slice": []interface{}{"$list", 0, max}}}},
	}
	cursor, err := deps.Mgo().Collection("comments").Aggregate(ctx, pipeline)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &lists)
	return
}
