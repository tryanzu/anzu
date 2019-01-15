package events

import (
	"time"

	ev "github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

type auditM struct {
	ID        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    bson.ObjectId `bson:"user_id" json:"user_id"`
	Related   string        `bson:"related" json:"related"`
	RelatedID bson.ObjectId `bson:"related_id" json:"related_id"`
	Reason    string        `bson:"reason" json:"reason"`
	Action    string        `bson:"action" json:"action"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
}

// Audit action log.
func audit(related string, id bson.ObjectId, action string, u ev.UserSign) {
	m := auditM{
		UserID:    u.UserID,
		Related:   related,
		RelatedID: id,
		Reason:    u.Reason,
		Action:    action,
		Created:   time.Now(),
	}
	err := deps.Container.Mgo().C("audits").Insert(&m)
	if err != nil {
		panic(err)
	}
}
