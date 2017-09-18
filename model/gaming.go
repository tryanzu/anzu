package model

import (
	"time"
)

type GamingRules struct {
	Updated time.Time   `json:"updated_at"`
	Rules   []RuleModel `json:"rules"`
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
