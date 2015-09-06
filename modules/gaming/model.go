package gaming

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type RulesModel struct {
	Updated time.Time    `json:"updated_at"`
	Rules   []RuleModel `json:"rules"`
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
	Slug          string        `bson:"slug" json:"slug"`
	Name          string        `bson:"name" json:"name"`
	Description   string        `bson:"description" json:"description"`
	Price         int           `bson:"price,omitempty" json:"price,omitempty"`
	RequiredBadge bson.ObjectId `bson:"required_badge,omitempty" json:"required_badge,omitempty"`
	RequiredLevel int           `bson:"required_level,omitempty" json:"required_level,omitempty"`
}