package feed

import (
	"github.com/tryanzu/core/modules/content"
	"github.com/tryanzu/core/modules/exceptions"
	//"github.com/tryanzu/core/modules/notifications"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/user"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
)

var lightPostFields bson.M = bson.M{"_id": 1, "title": 1, "slug": 1, "solved": 1, "lock": 1, "category": 1, "is_question": 1, "user_id": 1, "pinned": 1, "created_at": 1, "updated_at": 1, "type": 1, "content": 1}

type FeedModule struct {
	Errors       *exceptions.ExceptionsModule `inject:""`
	CacheService *goredis.Redis               `inject:""`
	User         *user.Module                 `inject:""`
	Content      *content.Module              `inject:""`
}

func (feed *FeedModule) SearchPosts(content string) ([]SearchPostModel, int) {

	posts := make([]SearchPostModel, 0)
	database := deps.Container.Mgo()

	// Fields to retrieve
	fields := bson.M{"_id": 1, "score": bson.M{"$meta": "textScore"}, "title": 1, "slug": 1, "solved": 1, "lock": 1, "category": 1, "user_id": 1, "pinned": 1, "created_at": 1, "updated_at": 1, "type": 1, "content": 1}
	query := bson.M{
		"$text": bson.M{"$search": content},
	}

	err := database.C("posts").Find(query).Select(fields).Sort("$textScore:score").Limit(10).All(&posts)

	if err != nil {
		panic(err)
	}

	count, err := database.C("posts").Find(query).Count()

	if err != nil {
		panic(err)
	}

	var (
		userIds []bson.ObjectId
		users   []user.UserSimple
	)

	for _, post := range posts {
		userIds = append(userIds, post.UserId)
	}

	err = database.C("users").Find(bson.M{"_id": bson.M{"$in": userIds}}).Select(user.UserSimpleFields).All(&users)
	if err != nil {
		panic(err)
	}

	usersMap := make(map[bson.ObjectId]interface{})

	for _, user := range users {
		usersMap[user.Id] = user
	}

	for index, post := range posts {
		if user, exists := usersMap[post.UserId]; exists {
			posts[index].User = user
		}
	}

	return posts, count
}

func (feed *FeedModule) Post(where interface{}) (post *Post, err error) {
	switch where.(type) {
	case bson.ObjectId, bson.M:
		var criteria = bson.M{"deleted_at": bson.M{"$exists": false}}

		switch where.(type) {
		case bson.ObjectId:
			criteria["_id"] = where.(bson.ObjectId)
		case bson.M:
			for k, v := range where.(bson.M) {
				criteria[k] = v
			}
		}

		// Use user feed reference to get the user and then create the user gaming instance
		err = deps.Container.Mgo().C("posts").Find(criteria).One(&post)
		if err != nil {
			err = exceptions.NotFound{Msg: "Invalid post id. Not found."}
			return
		}
	case *Post:
		post = where.(*Post)
	default:
		panic("Unkown argument")
	}

	post.SetDI(feed)
	return
}

func (feed *FeedModule) LightPost(post interface{}) (*LightPost, error) {

	switch post.(type) {
	case bson.ObjectId:

		scope := LightPostModel{}
		database := deps.Container.Mgo()

		// Use light post model
		err := database.C("posts").FindId(post.(bson.ObjectId)).Select(lightPostFields).One(&scope)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid post id. Not found."}
		}

		post_object := &LightPost{data: scope, di: feed}

		return post_object, nil

	default:
		panic("Unkown argument")
	}
}

func (feed *FeedModule) LightPosts(posts interface{}) ([]LightPostModel, error) {

	switch posts.(type) {
	case []bson.ObjectId:

		var list []LightPostModel

		database := deps.Container.Mgo()

		// Use light post model
		err := database.C("posts").Find(bson.M{"_id": bson.M{"$in": posts.([]bson.ObjectId)}}).Select(lightPostFields).All(&list)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid posts id. Not found."}
		}

		return list, nil

	case bson.M:

		var list []LightPostModel

		database := deps.Container.Mgo()

		// Use light post model
		err := database.C("posts").Find(posts.(bson.M)).Select(lightPostFields).All(&list)

		if err != nil {
			return nil, exceptions.NotFound{"Invalid posts criteria. Not found."}
		}

		return list, nil

	default:
		panic("Unkown argument")
	}
}

func (feed *FeedModule) GetComment(id bson.ObjectId) (comment *Comment, err error) {
	err = deps.Container.Mgo().C("comments").FindId(id).One(&comment)
	if err != nil {
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

	var ids []bson.ObjectId
	var comments []PostCommentModel

	for _, post := range list {

		// Generate the list of post id's
		ids = append(ids, post.Id)
	}

	database := deps.Container.Mgo()
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

	pipeline := database.C("posts").Pipe(pipeline_line)
	err := pipeline.All(&comments)

	if err != nil {
		panic(err)
	}

	assoc := map[bson.ObjectId]PostCommentModel{}

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

func (feed *FeedModule) TrueCommentCount(id bson.ObjectId) int {
	var count int

	database := deps.Container.Mgo()
	count, err := database.C("comments").Find(bson.M{"post_id": id}).Count()

	if err != nil {
		panic(err)
	}

	return count
}
