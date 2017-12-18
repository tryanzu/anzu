package notifications

import (
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"

	"fmt"
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

	database := deps.Container.Mgo()
	usr, err := self.User.Get(user_id)

	if err != nil {
		panic(fmt.Sprintf("Could not get user while notifying mention (user_id: %v, target_user: %v). It said: %v", user_id, target_user, err))
	}

	id, exists := parseableMeta["id"].(bson.ObjectId)

	if !exists {
		panic(fmt.Sprintf("ID does not exists in parseable meta (%v)", parseableMeta))
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
}
