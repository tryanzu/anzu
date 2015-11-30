package user

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type User struct {
	Id               bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	FirstName        string                 `bson:"first_name" json:"first_name"`
	LastName         string                 `bson:"last_name" json:"last_name"`
	UserName         string                 `bson:"username" json:"username"`
	UserNameSlug     string                 `bson:"username_slug" json:"username_slug"`
	NameChanges      int                    `bson:"name_changes" json:"name_changes"`
	Password         string                 `bson:"password" json:"-"`
	Step             int                    `bson:"step,omitempty" json:"step"`
	Email            string                 `bson:"email" json:"email,omitempty"`
	Categories       []bson.ObjectId        `bson:"categories,omitempty" json:"categories,omitempty"`
	Roles            []UserRole             `bson:"roles" json:"roles,omitempty"`
	Permissions      []string               `bson:"permissions" json:"permissions,omitempty"`
	Description      string                 `bson:"description" json:"description,omitempty"`
	Image            string                 `bson:"image" json:"image,omitempty"`
	Facebook         interface{}            `bson:"facebook,omitempty" json:"facebook,omitempty"`
	Notifications    interface{}            `bson:"notifications,omitempty" json:"notifications,omitempty"`
	Profile          map[string]interface{} `bson:"profile,omitempty" json:"profile,omitempty"`
	Gaming           UserGaming             `bson:"gaming,omitempty" json:"gaming,omitempty"`
	Stats            UserStats              `bson:"stats,omitempty" json:"stats,omitempty"`
	Version          string                 `bson:"version,omitempty" json:"version,omitempty"`
	ReferralCode     string                 `bson:"ref_code,omitempty" json:"ref_code"`
	VerificationCode string                 `bson:"ver_code,omitempty" json:"ver_code"`
	Validated        bool                   `bson:"validated" json:"validated"`
	Banned           bool                   `bson:"banned" json:"banned"`
	Created          time.Time              `bson:"created_at" json:"created_at"`
	Updated          time.Time              `bson:"updated_at" json:"updated_at"`
	Gamificated      time.Time              `bson:"gamificated_at" json:"gamificated_at"`

	// Runtime generated and not persisted in database
	Referrals ReferralsModel `json:"referrals,omitempty"`
}

type UserRole struct {
	Name       string          `bson:"name" json:"name"`
	Categories []bson.ObjectId `bson:"categories,omitempty" json:"categories,omitempty"`
}

type UserStats struct {
	Saw int `bson:"saw,omitempty" json:"saw,omitempty"`
}

type UserGaming struct {
	Swords  int         `bson:"swords" json:"swords"`
	Tribute int         `bson:"tribute" json:"tribute"`
	Shit    int         `bson:"shit" json:"shit"`
	Coins   int         `bson:"coins" json:"coins"`
	Level   int         `bson:"level" json:"level"`
	Badges  []UserBadge `bson:"badges" json:"badges"`
}

type UserBadge struct {
	Id   bson.ObjectId `bson:"id" json:"id"`
	Date time.Time     `bson:"date" json:"date"`
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

type UserLightModel struct {
	Id       bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Username string        `bson:"username" json:"username"`
	Email    string        `bson:"email" json:"email"`
	Image    string        `bson:"image" json:"image"`
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

type UserId struct {
	Id   bson.ObjectId `bson:"_id,omitempty" json:"id"`
}

type CheckinModel struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id"`
	UserId  bson.ObjectId `bson:"user_id" json:"user_id"`
	Address string        `bson:"client_ip" json:"client_ip"`
	Date    time.Time     `bson:"date" json:"date"`
}

type ReferralModel struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id"`
	OwnerId   bson.ObjectId `bson:"owner_id" json:"owner_id"`
	UserId    bson.ObjectId `bson:"user_id" json:"user_id"`
	Code      string        `bson:"ref_code" json:"ref_code"`
	Confirmed bool          `bson:"confirmed" json:"confirmed"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
	Updated   time.Time     `bson:"updated_at" json:"updated_at"`
}

type ReferralsModel struct {
	Count int              `json:"count"`
	List  []UserLightModel `json:"users"`
}

type ByCreatedAt []UserActivity

func (a ByCreatedAt) Len() int           { return len(a) }
func (a ByCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreatedAt) Less(i, j int) bool { return !a[i].Created.Before(a[j].Created) }
