package deps

import (
	"encoding/json"
	"io/ioutil"
)

var (
	ENV       string
	AppSecret string
)

func IgniteConfig(d Deps) (container Deps, err error) {
	gamingRules, err := ioutil.ReadFile("./gaming.json")
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
