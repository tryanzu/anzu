package deps

import (
	"encoding/json"
	"os"
)

var (
	ENV       string
	AppSecret string

	// SentryURL config
	SentryURL string
)

func IgniteConfig(d Deps) (container Deps, err error) {
	gamingRules, err := os.ReadFile("gaming.json")
	if err != nil {
		log.Error(err)
		return
	}

	err = json.Unmarshal(gamingRules, &d.GamingConfigProvider)
	if err != nil {
		log.Error(err)
		return
	}

	container = d
	return
}
