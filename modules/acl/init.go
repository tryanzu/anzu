package acl

import (
	"encoding/json"
	"io/ioutil"

	"github.com/mikespook/gorbac"
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

var LoadedACL *Module

type Module struct {
	Map         *gorbac.RBAC
	Rules       map[string]AclRole
	Permissions map[string]gorbac.Permission
}

func (module *Module) User(id bson.ObjectId) *User {

	var usr model.User
	database := deps.Container.Mgo()

	// Get the user using it's id
	err := database.C("users").FindId(id).One(&usr)

	if err != nil {
		panic(err)
	}

	user := &User{data: usr, acl: module}

	return user
}

func (refs *Module) CheckPermissions(roles []string, permission string) bool {

	for _, role := range roles {

		p, exists := refs.Permissions[permission]

		if exists {
			if refs.Map.IsGranted(role, p, nil) {
				// User's role is granted to do "permission"
				return true
			}
		}
	}

	return false
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
	module.Permissions = make(map[string]gorbac.Permission)

	for name, rules := range module.Rules {

		role := gorbac.NewStdRole(name)

		for _, p := range rules.Permissions {
			module.Permissions[p] = gorbac.NewStdPermission(p)
			role.Assign(module.Permissions[p])
		}

		// Populate map with permissions
		module.Map.Add(role)
	}

	for name, rules := range module.Rules {
		if len(rules.Inherits) > 0 {
			module.Map.SetParents(name, rules.Inherits)
		}
	}

	LoadedACL = module
	return module
}
