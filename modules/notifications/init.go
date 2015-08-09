package notifications

import (
	"github.com/fernandez14/spartangeek-blacker/mongo"
    "github.com/fernandez14/spartangeek-blacker/model"
    "github.com/fernandez14/spartangeek-blacker/modules/exceptions"
    "github.com/fernandez14/spartangeek-blacker/interfaces"
	"gopkg.in/mgo.v2/bson"
)
    
type NotificationsModule struct {
	Mongo 	    *mongo.Service 	                    `inject:""`
	Broadcaster interfaces.NotificationBroadcaster  `inject:"Notifications"`
	Errors      *exceptions.ExceptionsModule	    `inject:""`
}

type MentionParseObject struct {
    Type string
    RelatedNested string
    Content string
    Title string
    Author bson.ObjectId
    Post model.Post
}