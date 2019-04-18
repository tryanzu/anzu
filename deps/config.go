package deps

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/olebedev/config"
	"github.com/subosito/gotenv"
)

func IgniteConfig(d Deps) (container Deps, err error) {
	gotenv.Load()
	envfile := os.Getenv("ENV_FILE")
	if envfile == "" {
		if curr, err := os.Getwd(); err == nil {
			envfile = curr + "/env.json"
		}
	}

	parsed, err := config.ParseJsonFile(envfile)
	if err != nil {
		return
	}

	// Load other files with important config.
	gaming, err := parsed.String("application.gaming")
	if err != nil {
		return
	}

	gamingRules, err := ioutil.ReadFile(gaming)
	if err != nil {
		return
	}

	err = json.Unmarshal(gamingRules, &d.GamingConfigProvider)
	if err != nil {
		return
	}

	d.ConfigProvider = parsed
	container = d
	container.Log().Debugf("[CONFIG] Read %s", envfile)
	return
}
