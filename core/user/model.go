package user

import (
	"time"

	"gopkg.in/mgo.v2/bson"
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
	Categories    []bson.ObjectId        `bson:"categories,omitempty" json:"categories,omitempty"`
	Roles         []UserRole             `bson:"roles" json:"roles,omitempty"`
	Permissions   []string               `bson:"permissions" json:"permissions,omitempty"`
	Description   string                 `bson:"description" json:"description,omitempty"`
	Image         string                 `bson:"image" json:"image"`
	Facebook      interface{}            `bson:"facebook,omitempty" json:"facebook,omitempty"`
	Notifications interface{}            `bson:"notifications,omitempty" json:"notifications,omitempty"`
	Profile       map[string]interface{} `bson:"profile,omitempty" json:"profile,omitempty"`
	Gaming        Gaming                 `bson:"gaming,omitempty" json:"gaming,omitempty"`
	//Stats         UserStats              `bson:"stats,omitempty" json:"stats,omitempty"`
	Version          string `bson:"version,omitempty" json:"version,omitempty"`
	Validated        bool   `bson:"validated" json:"validated"`
	VerificationCode string `bson:"ver_code,omitempty" json:"-"`

	Warnings    int        `bson:"warnings" json:"-"`
	ConfirmSent *time.Time `bson:"confirm_sent_at" json:"-"`
	Created     time.Time  `bson:"created_at" json:"created_at"`
	Updated     time.Time  `bson:"updated_at" json:"updated_at"`

	Banned       bool       `bson:"banned" json:"-"`
	BannedAt     *time.Time `bson:"banned_at" json:"-"`
	BannedUntil  *time.Time `bson:"banned_until" json:"-"`
	BannedReason *string    `bson:"banned_re" json:"-"`
	BannedTimes  int        `bson:"banned_times" json:"-"`
}

type UserRole struct {
	Name string `bson:"name" json:"name"`
}

func (usr User) HasRole(roles ...string) bool {
	for _, role := range usr.Roles {
		for _, validRole := range roles {
			if validRole == role.Name {
				return true
			}
		}
	}
	return false
}

type Gaming struct {
	Swords  int `bson:"swords" json:"swords"`
	Tribute int `bson:"tribute" json:"tribute"`
	Shit    int `bson:"shit" json:"shit"`
	Coins   int `bson:"coins" json:"coins"`
	Level   int `bson:"level" json:"level"`
}

type Users []User

type UsersSet struct {
	List  Users `json:"list"`
	Count int   `json:"count"`
}

func (list Users) Map() map[bson.ObjectId]User {
	m := make(map[bson.ObjectId]User, len(list))
	for _, item := range list {
		m[item.Id] = item
	}

	return m
}

func (list Users) UpdateCache(d deps) (err error) {
	for _, u := range list {
		err = d.LedisDB().Set([]byte("user:"+u.Id.Hex()+":names"), []byte(u.UserName))
		if err != nil {
			return
		}
	}
	return
}

type RecoveryToken struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Token   string        `bson:"token" json:"token"`
	UserId  bson.ObjectId `bson:"user_id" json:"user_id"`
	Used    bool          `bson:"used" json:"used"`
	Created time.Time     `bson:"created_at" json:"created_at"`
	Updated time.Time     `bson:"updated_at" json:"updated_at"`
}

type status string
type category string

const (
	ACTIVE   status = "active"
	PENDING  status = "pending"
	REJECTED status = "rejected"
	REVOKED  status = "revoked"
)

// Ban represents a ban sent by a user.
type Ban struct {
	ID        bson.ObjectId  `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    bson.ObjectId  `bson:"user_id" json:"user_id"`
	RelatedTo string         `bson:"related_to" json:"related_to"`
	RelatedID *bson.ObjectId `bson:"related_id,omitempty" json:"related_id,omitempty"`
	Content   string         `bson:"content" json:"content"`
	Status    status         `bson:"status" json:"status"`
	Reason    string         `bson:"reason" json:"reason"`
	Created   time.Time      `bson:"created_at" json:"created_at"`
	Updated   time.Time      `bson:"updated_at" json:"updated_at"`
	Until     time.Time      `bson:"until" json:"until"`
	Deleted   *time.Time     `bson:"deleted_at,omitempty" json:"-"`
}
