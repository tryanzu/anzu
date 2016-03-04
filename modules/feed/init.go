package feed

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/modules/search"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"

	"fmt"
)

var lightPostFields bson.M = bson.M{"_id": 1, "title": 1, "slug": 1, "solved": 1, "lock": 1, "category": 1, "user_id": 1, "pinned": 1, "created_at": 1, "updated_at": 1, "type": 1, "content": 1}

type FeedModule struct {
	Mongo        *mongo.Service               `inject:""`
	Errors       *exceptions.ExceptionsModule `inject:""`
	CacheService *goredis.Redis               `inject:""`
	Search       *search.Module               `inject:""`
	User         *user.Module                 `inject:""`
}

func (module *FeedModule) SearchPosts(content string) []SearchPostModel {

	var posts []SearchPostModel
	database := module.Mongo.Database

	// Make the search a phrase search
	content = "\"" + content + "\""

	// Fields to retrieve
	fields := bson.M{"_id": 1, "score": bson.M{"$meta": "textScore"}, "title": 1, "slug": 1, "solved": 1, "lock": 1, "category": 1, "user_id": 1, "pinned": 1, "created_at": 1, "updated_at": 1, "type": 1, "content": 1}
	query := bson.M{
		"$text": bson.M{"$search": content},
	}

	err := database.C("posts").Find(query).Select(fields).Sort("$textScore:score").Limit(10).All(&posts)

	if err != nil {
		panic(err)
	}

	fmt.Println(len(posts))
	fmt.Println(content)

	var users_id []bson.ObjectId
	var users []user.UserSimple

	for _, post := range posts {
		users_id = append(users_id, post.UserId)
	}

	err = database.C("users").Find(bson.M{"_id": bson.M{"$in": users_id}}).Select(user.UserSimpleFields).All(&users)
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

	return posts
}

func (self *FeedModule) Post(post interface{}) (*Post, error) {

	module := self

	switch post.(type) {
	case bson.ObjectId:

		this := model.Post{}
		database := self.Mongo.Database

		// Use user module reference to get the user and then create the user gaming instance
		err := database.C("posts").FindId(post.(bson.ObjectId)).One(&this)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid post id. Not found."}
		}

		post_object := &Post{data: this, di: module}

		return post_object, nil

	case model.Post:

		this := post.(model.Post)
		post_object := &Post{data: this, di: module}

		return post_object, nil

	default:
		panic("Unkown argument")
	}
}

func (module *FeedModule) LightPost(post interface{}) (*LightPost, error) {

	switch post.(type) {
	case bson.ObjectId:

		scope := LightPostModel{}
		database := module.Mongo.Database

		// Use light post model 
		err := database.C("posts").FindId(post.(bson.ObjectId)).Select(bson.M{"_id": 1, "title": 1, "slug": 1, "category": 1, "user_id": 1, "lock": 1, "pinned": 1, "created_at": 1, "updated_at": 1, "type": 1, "content": 1}).One(&scope)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid post id. Not found."}
		}

		post_object := &LightPost{data: scope, di: module}

		return post_object, nil

	default:
		panic("Unkown argument")
	}
}

func (module *FeedModule) LightPosts(posts interface{}) ([]LightPostModel, error) {

	switch posts.(type) {
	case []bson.ObjectId:

		var list []LightPostModel

		database := module.Mongo.Database

		// Use light post model 
		err := database.C("posts").Find(bson.M{"_id": bson.M{"$in": posts.([]bson.ObjectId)}}).Select(lightPostFields).All(&list)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid posts id. Not found."}
		}

		return list, nil

	case bson.M:

		var list []LightPostModel

		database := module.Mongo.Database

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

func (module *FeedModule) FulfillBestAnswer(list []LightPostModel) []LightPostModel {

	var ids []bson.ObjectId
	var comments []PostCommentModel

	for _, post := range list {

		// Generate the list of post id's
		ids = append(ids, post.Id)
	}

	database := module.Mongo.Database
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

func (module *FeedModule) TrueCommentCount(id bson.ObjectId) int {

	var count PostCommentCountModel
	database := module.Mongo.Database

	pipe := database.C("posts").Pipe([]bson.M{
		{"$match": bson.M{"_id": id}},
		{"$project": bson.M{"count": bson.M{"$size": "$comments.set"}}},
	})

	err := pipe.One(&count)

	if err != nil {
		return 0
	}

	return count.Count
}

func (module *FeedModule) Posts(limit, offset int) List {

	list := List{
		module: module,
		limit: limit,
		offset: offset,
	}

	return list
}