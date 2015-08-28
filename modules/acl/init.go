package acl

import (
	"encoding/json"
	"github.com/mikespook/gorbac"
	"io/ioutil"
)

type Module struct {
	Rules map[string]AclRole
	Map   *gorbac.Rbac
}

func Boot(file string) *Module {

	module := &Module{}
	rules_data, err := ioutil.ReadFile(file)

	if err != nil {
		panic(err)
	}

	// Unmarshal file with gaming rules
	if err := json.Unmarshal(rules_data, &module.Rules); err != nil {
		panic(err)
	}

	module.Map = gorbac.New()

	for name, rules := range module.Rules {

		// Populate map with permissions
		module.Map.Add(name, rules.Permissions, rules.Inherits)
	}

	return module
}
