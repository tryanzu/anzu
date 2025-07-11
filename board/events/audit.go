package events

import (
	"context"
	"time"

	ev "github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/deps"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type auditM struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Related   string             `bson:"related" json:"related"`
	RelatedID primitive.ObjectID `bson:"related_id" json:"related_id"`
	Reason    string             `bson:"reason" json:"reason"`
	Action    string             `bson:"action" json:"action"`
	Created   time.Time          `bson:"created_at" json:"created_at"`
}

// Audit action log.
func audit(related string, id primitive.ObjectID, action string, u ev.UserSign) {
	m := auditM{
		UserID:    u.UserID,
		Related:   related,
		RelatedID: id,
		Reason:    u.Reason,
		Action:    action,
		Created:   time.Now(),
	}
	_, err := deps.Container.Mgo().Collection("audits").InsertOne(context.Background(), &m)
	if err != nil {
		panic(err)
	}
}
