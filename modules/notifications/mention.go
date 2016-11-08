package notifications

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"gopkg.in/mgo.v2/bson"

	"fmt"
	"strconv"
	"time"
)

type Mention struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	UserId    bson.ObjectId `bson:"user_id" json:"user_id"` // Mentioned
	FromId    bson.ObjectId `bson:"from_id" json:"from_id"`
	Related   string        `bson:"related" json:"related"`
	RelatedId bson.ObjectId `bson:"related_id" json:"related_id"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
	Updated   time.Time     `bson:"updated_at" json:"updated_at"`
}

func (self *NotificationsModule) Mention(parseableMeta map[string]interface{}, user_id, target_user bson.ObjectId) {

	defer self.Errors.Recover()

	database := self.Mongo.Database
	usr, err := self.User.Get(user_id)

	if err != nil {
		panic(fmt.Sprintf("Could not get user while notifying mention (user_id: %v, target_user: %v). It said: %v", user_id, target_user, err))
	}

	position, exists := parseableMeta["position"].(int)

	if !exists {
		panic(fmt.Sprintf("Position does not exists in parseable meta (%v)", parseableMeta))
	}

	id, exists := parseableMeta["id"].(bson.ObjectId)

	if !exists {
		panic(fmt.Sprintf("ID does not exists in parseable meta (%v)", parseableMeta))
	}

	post, exists := parseableMeta["post"].(map[string]interface{})

	if !exists {
		panic(fmt.Sprintf("Post does not exists in parseable meta (%v)", parseableMeta))
	}

	post_id, exists := post["id"].(bson.ObjectId)

	if !exists {
		panic(fmt.Sprintf("post_id does not exists in parseable meta (%v)", parseableMeta))
	}

	post_slug, exists := post["slug"].(string)

	if !exists {
		panic(fmt.Sprintf("post_slug does not exists in parseable meta (%v)", parseableMeta))
	}

	post_title, exists := post["title"].(string)

	if !exists {
		panic(fmt.Sprintf("post_title does not exists in parseable meta (%v)", parseableMeta))
	}

	mention := Mention{
		UserId:    target_user,
		Related:   "comment",
		RelatedId: id,
		FromId:    usr.Data().Id,
		Created:   time.Now(),
		Updated:   time.Now(),
	}

	err = database.C("mentions").Insert(mention)

	if err != nil {
		panic(err)
	}

	notification := model.UserFirebaseNotification{
		UserId:       target_user,
		RelatedId:    post_id,
		RelatedExtra: post_slug,
		Position:     strconv.Itoa(position),
		Username:     usr.Name(),
		Text:         post_title,
		Related:      "mention",
		Seen:         false,
		Image:        usr.Data().Image,
		Created:      time.Now(),
		Updated:      time.Now(),
	}

	broadcaster := self.Broadcaster
	broadcaster.Send(notification)
}
