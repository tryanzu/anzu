package user

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type User struct {
	Id           bson.ObjectId   `bson:"_id,omitempty" json:"id"`
	FirstName    string          `bson:"first_name" json:"first_name"`
	LastName     string          `bson:"last_name" json:"last_name"`
	UserName     string          `bson:"username" json:"username"`
	UserNameSlug string          `bson:"username_slug" json:"username_slug"`
	NameChanges  int             `bson:"name_changes" json:"name_changes"`
	Password     string          `bson:"password" json:"-"`
	Step         int             `bson:"step,omitempty" json:"step"`
	Email        string          `bson:"email" json:"email,omitempty"`
	Categories   []bson.ObjectId `bson:"categories,omitempty" json:"categories,omitempty"`
	//Roles         []UserRole             `bson:"roles" json:"roles,omitempty"`
	Permissions   []string               `bson:"permissions" json:"permissions,omitempty"`
	Description   string                 `bson:"description" json:"description,omitempty"`
	Image         string                 `bson:"image" json:"image,omitempty"`
	Facebook      interface{}            `bson:"facebook,omitempty" json:"facebook,omitempty"`
	Notifications interface{}            `bson:"notifications,omitempty" json:"notifications,omitempty"`
	Profile       map[string]interface{} `bson:"profile,omitempty" json:"profile,omitempty"`
	//Gaming        UserGaming             `bson:"gaming,omitempty" json:"gaming,omitempty"`
	//Stats         UserStats              `bson:"stats,omitempty" json:"stats,omitempty"`
	Version   string `bson:"version,omitempty" json:"version,omitempty"`
	Validated bool   `bson:"validated" json:"validated"`

	Warnings int       `bson:"warnings" json:"-"`
	Created  time.Time `bson:"created_at" json:"created_at"`
	Updated  time.Time `bson:"updated_at" json:"updated_at"`
}
