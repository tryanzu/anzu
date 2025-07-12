package feed

import (
	"context"

	"github.com/tryanzu/core/modules/content"
	"github.com/tryanzu/core/modules/exceptions"

	//"github.com/tryanzu/core/modules/notifications"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/user"
	"github.com/xuyu/goredis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var lightPostFields bson.M = bson.M{"_id": 1, "title": 1, "slug": 1, "solved": 1, "lock": 1, "category": 1, "is_question": 1, "user_id": 1, "pinned": 1, "created_at": 1, "updated_at": 1, "type": 1, "content": 1}

type FeedModule struct {
	Errors       *exceptions.ExceptionsModule `inject:""`
	CacheService *goredis.Redis               `inject:""`
	User         *user.Module                 `inject:""`
	Content      *content.Module              `inject:""`
}

func (feed *FeedModule) Post(where interface{}) (post *Post, err error) {
	ctx := context.Background()
	switch where.(type) {
	case primitive.ObjectID, bson.M:
		var criteria = bson.M{"deleted_at": bson.M{"$exists": false}}

		switch where.(type) {
		case primitive.ObjectID:
			criteria["_id"] = where.(primitive.ObjectID)
		case bson.M:
			for k, v := range where.(bson.M) {
				criteria[k] = v
			}
		}

		// Use user feed reference to get the user and then create the user gaming instance
		database := deps.Container.Mgo()
		collection := database.Collection("posts")
		err = collection.FindOne(ctx, criteria).Decode(&post)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				err = exceptions.NotFound{Msg: "Invalid post id. Not found."}
			}
			return
		}
	case *Post:
		post = where.(*Post)
	default:
		panic("Unknown argument")
	}

	post.SetDI(feed)
	return
}

func (feed *FeedModule) LightPost(post interface{}) (*LightPost, error) {
	ctx := context.Background()
	switch post.(type) {
	case primitive.ObjectID:
		scope := LightPostModel{}
		database := deps.Container.Mgo()
		collection := database.Collection("posts")

		// Use light post model
		opts := options.FindOne().SetProjection(lightPostFields)
		err := collection.FindOne(ctx, bson.M{"_id": post.(primitive.ObjectID)}, opts).Decode(&scope)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, exceptions.NotFound{"Invalid post id. Not found."}
			}
			return nil, err
		}

		post_object := &LightPost{data: scope, di: feed}
		return post_object, nil

	default:
		panic("Unknown argument")
	}
}

func (feed *FeedModule) LightPosts(posts interface{}) ([]LightPostModel, error) {
	ctx := context.Background()
	switch posts.(type) {
	case []primitive.ObjectID:
		postIDs := posts.([]primitive.ObjectID)
		if len(postIDs) == 0 {
			return []LightPostModel{}, nil
		}

		var list []LightPostModel
		database := deps.Container.Mgo()
		collection := database.Collection("posts")

		// Use light post model
		filter := bson.M{"_id": bson.M{"$in": postIDs}}
		opts := options.Find().SetProjection(lightPostFields)
		cursor, err := collection.Find(ctx, filter, opts)
		if err != nil {
			return nil, exceptions.NotFound{"Invalid posts id. Not found."}
		}
		defer cursor.Close(ctx)

		err = cursor.All(ctx, &list)
		if err != nil {
			return nil, err
		}
		return list, nil

	case bson.M:
		var list []LightPostModel
		database := deps.Container.Mgo()
		collection := database.Collection("posts")

		// Use light post model
		opts := options.Find().SetProjection(lightPostFields)
		cursor, err := collection.Find(ctx, posts.(bson.M), opts)
		if err != nil {
			return nil, exceptions.NotFound{"Invalid posts criteria. Not found."}
		}
		defer cursor.Close(ctx)

		err = cursor.All(ctx, &list)
		if err != nil {
			return nil, err
		}
		return list, nil

	default:
		panic("Unknown argument")
	}
}

func (feed *FeedModule) GetComment(id primitive.ObjectID) (comment *Comment, err error) {
	ctx := context.Background()
	database := deps.Container.Mgo()
	collection := database.Collection("comments")
	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&comment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, exceptions.NotFound{"Comment not found"}
		}
		return
	}

	post, err := feed.Post(comment.PostId)
	if err != nil {
		return nil, err
	}

	comment.SetDI(post)
	return
}

func (feed *FeedModule) FulfillBestAnswer(list []LightPostModel) []LightPostModel {
	ctx := context.Background()
	var ids []primitive.ObjectID
	var comments []PostCommentModel

	for _, post := range list {
		// Generate the list of post id's
		ids = append(ids, post.Id)
	}

	database := deps.Container.Mgo()
	collection := database.Collection("posts")
	pipeline_line := []bson.M{
		{
			"$match": bson.M{"_id": bson.M{"$in": ids}, "solved": true},
		},
		{
			"$unwind": "$comments.set",
		},
		{
			"$match": bson.M{"comments.set.chosen": true},
		},
		{
			"$project": bson.M{"comment": "$comments.set"},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline_line)
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &comments)
	if err != nil {
		panic(err)
	}

	assoc := map[primitive.ObjectID]PostCommentModel{}

	for _, comment := range comments {
		assoc[comment.Id] = comment
	}

	for index, post := range list {
		if comment, exists := assoc[post.Id]; exists {
			list[index].BestAnswer = &comment.Comment
		}
	}

	return list
}

func (feed *FeedModule) TrueCommentCount(id primitive.ObjectID) int {
	ctx := context.Background()
	database := deps.Container.Mgo()
	collection := database.Collection("comments")
	count, err := collection.CountDocuments(ctx, bson.M{"post_id": id})
	if err != nil {
		panic(err)
	}

	return int(count)
}
