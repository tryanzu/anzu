package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type User struct {
	Id            bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	FirstName     string                 `bson:"first_name" json:"first_name"`
	LastName      string                 `bson:"last_name" json:"last_name"`
	UserName      string                 `bson:"username" json:"username"`
	UserNameSlug  string                 `bson:"username_slug" json:"username_slug"`
	NameChanges   int                    `bson:"name_changes" json:"name_changes"`
	Password      string                 `bson:"password" json:"-"`
	Step          int                    `bson:"step,omitempty" json:"step"`
	Email         string                 `bson:"email" json:"email,omitempty"`
	Roles         []string               `bson:"roles" json:"roles,omitempty"`
	Permissions   []string               `bson:"permissions" json:"permissions,omitempty"`
	Description   string                 `bson:"description" json:"description,omitempty"`
	Image         string                 `bson:"image" json:"image,omitempty"`
	Facebook      interface{}            `bson:"facebook,omitempty" json:"facebook,omitempty"`
	Notifications interface{}            `bson:"notifications,omitempty" json:"notifications,omitempty"`
	Profile       map[string]interface{} `bson:"profile,omitempty" json:"profile,omitempty"`
	Stats         UserStats              `bson:"stats,omitempty" json:"stats,omitempty"`
	Version       string                 `bson:"version,omitempty" json:"version,omitempty"`
	Created       time.Time              `bson:"created_at" json:"created_at"`
	Updated       time.Time              `bson:"updated_at" json:"updated_at"`
}

type UserStats struct {
	Saw int `bson:"saw,omitempty" json:"saw,omitempty"`
}

type UserToken struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	UserId  bson.ObjectId `bson:"user_id" json:"user_id"`
	Token   string        `bson:"token" json:"token"`
	Closed  bool          `bson:"closed,omitempty" json"closed,omitempty"`
	Created time.Time     `bson:"created_at" json:"created_at"`
	Updated time.Time     `bson:"updated_at" json:"updated_at"`
}

type UserPc struct {
	Type string `bson:"type" json:"type"`
}

type UserFollowing struct {
	Id            bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Follower      bson.ObjectId `bson:"follower,omitempty" json:"follower"`
	Following     bson.ObjectId `bson:"following,omitempty" json:"following"`
	Notifications bool          `bson:"notifications,omitempty" json:"notifications"`
	Created       time.Time     `bson:"created_at" json:"created_at"`
}

type UserActivity struct {
	Title     string            `json:"title"`
	Directive string            `json:"directive"`
	Content   string            `json:"content"`
	Author    map[string]string `json:"user"`
	Created   time.Time         `json:"created_at"`
}

type UserProfileForm struct {
	UserName string `json:"username,omitempty"`
}

type UserRegisterForm struct {
	UserName string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

type UserSubscribe struct {
	Id       bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Category string        `bson:"category" json:"category"`
	Email    string        `bson:"email" json:"email"`
}

type UserSubscribeForm struct {
	Category string `json:"category" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

type ByCreatedAt []UserActivity

func (a ByCreatedAt) Len() int           { return len(a) }
func (a ByCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreatedAt) Less(i, j int) bool { return !a[i].Created.Before(a[j].Created) }
