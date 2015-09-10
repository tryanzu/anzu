package gaming

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type RulesModel struct {
	Updated time.Time    `json:"updated_at"`
	Rules   []RuleModel  `json:"rules"`
	Badges  []BadgeModel `json:"badges,omitempty"`
}

type RuleModel struct {
	Level   int    `json:"level"`
	Name    string `json:"name"`
	Start   int    `json:"swords_start"`
	End     int    `json:"swords_end"`
	Tribute int    `json:"tribute"`
	Shit    int    `json:"shit"`
	Coins   int    `json:"coins"`
}

type BadgeModel struct {
	Id            bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Type          string        `bson:"type" json:"type"`
	TypeLabel     string        `bson:"type_label" json:"type_label"`
	Slug          string        `bson:"slug" json:"slug"`
	Name          string        `bson:"name" json:"name"`
	Description   string        `bson:"description" json:"description"`
	Coins         int           `bson:"coins,omitempty" json:"coins,omitempty"`
	RequiredBadge bson.ObjectId `bson:"required_badge,omitempty" json:"required_badge,omitempty"`
	RequiredLevel int           `bson:"required_level,omitempty" json:"required_level,omitempty"`
	Avaliable     bool          `bson:"available" json:"available"`
}

type RankingModel struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	UserId    bson.ObjectId `bson:"user_id" json:"user_id"`
	Badges    int           `bson:"badges" json:"badges"`
	Swords    int           `bson:"swords" json:"swords"`
	Coins     int           `bson:"coins" json:"coins"`
	Position  RankingPositionModel `bson:"position" json:"position"`
	Before    RankingPositionModel `bson:"before" json:"before"`	
	User      RankingUserModel     `json:"user,omitempty"`
	Created   time.Time            `bson:"created_at" json:"created_at"`
}

type RankingPositionModel struct {
	Wealth int `bson:"wealth" json:"wealth"`
	Swords int `bson:"swords" json:"swords"`
	Badges int `bson:"badges" json:"badges"`
}

type RankingUserModel struct {
	Id       bson.ObjectId `bson:"_id,omitempty" json:"id"`
	UserName string `bson:"username" json:"username"`
	Image    string `bson:"image" json:"image,omitempty"`
}

type RankBySwords []RankingModel

func (a RankBySwords) Len() int           { return len(a) }
func (a RankBySwords) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RankBySwords) Less(i, j int) bool { return a[i].Swords > a[j].Swords }

type RankByCoins []RankingModel

func (a RankByCoins) Len() int           { return len(a) }
func (a RankByCoins) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RankByCoins) Less(i, j int) bool { return a[i].Coins > a[j].Coins }

type RankByBadges []RankingModel

func (a RankByBadges) Len() int           { return len(a) }
func (a RankByBadges) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RankByBadges) Less(i, j int) bool { return a[i].Badges > a[j].Badges }