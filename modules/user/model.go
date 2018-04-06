package user

import (
	"gopkg.in/mgo.v2/bson"

	"time"
)

type User struct {
	Id          bson.ObjectId `bson:"_id,omitempty" json:"id"`
	UserName    string        `bson:"username" json:"username"`
	NameChanges int           `bson:"name_changes" json:"name_changes"`
	Description string        `bson:"description" json:"description,omitempty"`
	Image       string        `bson:"image" json:"image,omitempty"`
	Roles       []UserRole    `bson:"roles" json:"roles,omitempty"`
	Permissions []string      `bson:"permissions" json:"permissions,omitempty"`

	Profile map[string]interface{} `bson:"profile,omitempty" json:"profile,omitempty"`
	Gaming  UserGaming             `bson:"gaming,omitempty" json:"gaming,omitempty"`

	Version   string    `bson:"version,omitempty" json:"version,omitempty"`
	Validated bool      `bson:"validated" json:"validated"`
	Banned    bool      `bson:"banned" json:"banned"`
	Created   time.Time `bson:"created_at" json:"created_at"`
}

type UserPrivate struct {
	User             `bson:",inline"`
	Password         string          `bson:"password,omitempty" json:"-"`
	Step             int             `bson:"step,omitempty" json:"step"`
	Notifications    int             `bson:"notifications,omitempty" json:"notifications"`
	Email            string          `bson:"email,omitempty" json:"email,omitempty"`
	Categories       []bson.ObjectId `bson:"categories,omitempty" json:"categories,omitempty"`
	Facebook         interface{}     `bson:"facebook,omitempty" json:"facebook,omitempty"`
	Stats            UserStats       `bson:"stats,omitempty" json:"stats,omitempty"`
	ReferralCode     string          `bson:"ref_code,omitempty" json:"ref_code"`
	VerificationCode string          `bson:"ver_code,omitempty" json:"ver_code"`
	SessionId        string          `bson:"-" json:"session_id"`
	Duplicates       []bson.ObjectId `bson:"duplicates" json:"-"`
	ConfirmationSent *time.Time      `bson:"confirm_sent_at" json:"-"`
	Updated          time.Time       `bson:"updated_at" json:"updated_at"`
	Gamificated      time.Time       `bson:"gamificated_at" json:"gamificated_at"`
}

type UserSimple struct {
	Id           bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Roles        []UserRole    `bson:"roles" json:"roles,omitempty"`
	UserName     string        `bson:"username" json:"username"`
	UserNameSlug string        `bson:"username_slug" json:"username_slug"`
	Gaming       UserGaming    `bson:"gaming,omitempty" json:"gaming,omitempty"`
	Image        string        `bson:"image" json:"image,omitempty"`
	Description  string        `bson:"description" json:"description"`

	Country     string `bson:"country,omitempty" json:"country"`
	OriginId    string `bson:"origin_id,omitempty" json:"origin_id"`
	BattlenetId string `bson:"battlenet_id,omitempty" json:"battlenet_id"`
	SteamId     string `bson:"steam_id,omitempty" json:"steam_id"`

	Validated bool      `bson:"validated" json:"validated"`
	Created   time.Time `bson:"created_at" json:"created_at"`
}

type UserRecoveryToken struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Token   string        `bson:"token" json:"token"`
	UserId  bson.ObjectId `bson:"user_id" json:"user_id"`
	Used    bool          `bson:"used" json:"used"`
	Created time.Time     `bson:"created_at" json:"created_at"`
	Updated time.Time     `bson:"updated_at" json:"updated_at"`
}

var UserSimpleFields bson.M = bson.M{"id": 1, "username": 1, "description": 1, "username_slug": 1, "country": 1, "origin_id": 1, "battlenet_id": 1, "steam_id": 1, "image": 1, "gaming.level": 1, "gaming.swords": 1, "roles": 1, "validated": 1, "created_at": 1}
var UserBasicFields bson.M = bson.M{"id": 1, "username": 1, "description": 1, "facebook": 1, "country": 1, "origin_id": 1, "battlenet_id": 1, "steam_id": 1, "email": 1, "validated": 1, "banned": 1, "username_slug": 1, "image": 1, "gaming": 1, "created_at": 1, "updated_at": 1}

type UserBasic struct {
	Id           bson.ObjectId `bson:"_id,omitempty" json:"id"`
	UserName     string        `bson:"username" json:"username"`
	UserNameSlug string        `bson:"username_slug" json:"username_slug"`
	Roles        []UserRole    `bson:"roles" json:"roles,omitempty"`
	Image        string        `bson:"image" json:"image,omitempty"`
	Description  string        `bson:"description" json:"description"`
	Email        string        `bson:"email" json:"email,omitempty"`
	Facebook     interface{}   `bson:"facebook,omitempty" json:"facebook,omitempty"`
	Gaming       UserGaming    `bson:"gaming,omitempty" json:"gaming,omitempty"`

	Country     string `bson:"country,omitempty" json:"country"`
	OriginId    string `bson:"origin_id,omitempty" json:"origin_id"`
	BattlenetId string `bson:"battlenet_id,omitempty" json:"battlenet_id"`
	SteamId     string `bson:"steam_id,omitempty" json:"steam_id"`

	Validated bool      `bson:"validated" json:"validated"`
	Banned    bool      `bson:"banned" json:"banned"`
	Created   time.Time `bson:"created_at" json:"created_at"`
	Updated   time.Time `bson:"updated_at" json:"updated_at"`
}

func (u UserBasic) ToSimple() UserSimple {
	return UserSimple{u.Id, u.Roles, u.UserName, u.UserNameSlug, u.Gaming, u.Image, u.Description, u.Country, u.OriginId, u.BattlenetId, u.SteamId, u.Validated, u.Created}
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
	Tribute int         `bson:"tribute,omitempty" json:"tribute"`
	Shit    int         `bson:"shit,omitempty" json:"shit"`
	Coins   int         `bson:"coins,omitempty" json:"coins"`
	Level   int         `bson:"level" json:"level"`
	Badges  []UserBadge `bson:"badges,omitempty" json:"badges,omitempty"`
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
	Id bson.ObjectId `bson:"_id,omitempty" json:"id"`
}

type ViewModel struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id"`
	UserId    bson.ObjectId `bson:"user_id" json:"user_id"`
	Related   string        `bson:"related" json:"related"`
	RelatedId bson.ObjectId `bson:"related_id" json:"related_id"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
}

type CheckinModel struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id"`
	UserId  bson.ObjectId `bson:"user_id" json:"user_id"`
	Address string        `bson:"client_ip" json:"client_ip"`
	Date    time.Time     `bson:"date" json:"date"`
}

type ByCreatedAt []UserActivity

func (a ByCreatedAt) Len() int           { return len(a) }
func (a ByCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreatedAt) Less(i, j int) bool { return !a[i].Created.Before(a[j].Created) }
