package deps

import (
	"os"

	"github.com/olebedev/config"
)

func IgniteConfig(container Deps) (Deps, error) {
	envfile := os.Getenv("ENV_FILE")
	if envfile == "" {
		curr, err := os.Getwd()
		if err != nil {
			return container, err
		}

		envfile = curr + "/env.json"
	}

	conf, err := config.ParseJsonFile(envfile)
	if err != nil {
		return container, err
	}

	container.Config = conf
	container.Logger.Debugf("Configuration ignited using %s", envfile)
	return container, nil
}
