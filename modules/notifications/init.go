package notifications

import (
	"github.com/cosn/firebase"
	"github.com/fernandez14/spartangeek-blacker/mongo"
    "github.com/fernandez14/spartangeek-blacker/model"
    "github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"gopkg.in/mgo.v2/bson"
)
    
type NotificationsModule struct {
	Mongo 	    *mongo.Service 	        `inject:""`
	Firebase    *firebase.Client        `inject:""`
	Errors      *exceptions.ExceptionsModule	`inject:""`
}

type MentionParseObject struct {
    Type string
    RelatedNested string
    Content string
    Title string
    Author bson.ObjectId
    Post model.Post
}