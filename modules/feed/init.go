package feed

import (
	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
)

type FeedModule struct {
	Mongo        *mongo.Service               `inject:""`
	Errors       *exceptions.ExceptionsModule `inject:""`
	CacheService *goredis.Redis               `inject:""`
	Algolia      *algoliasearch.Index         `inject:""`
	User         *user.Module                 `inject:""`
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

func (module *FeedModule) Posts(limit, offset int) List {

	list := List{
		module: module,
		limit: limit,
		offset: offset,
	}

	return list
}