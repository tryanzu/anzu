package acl

import (
	"encoding/json"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/mikespook/gorbac"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
)

type Module struct {
	Map   *gorbac.RBAC
	Mongo *mongo.Service `inject:""`
	Rules map[string]AclRole
	Permissions map[string]gorbac.Permission
}

func (module *Module) User(id bson.ObjectId) *User {

	var usr model.User
	database := module.Mongo.Database

	// Get the user using it's id
	err := database.C("users").FindId(id).One(&usr)

	if err != nil {
		panic(err)
	}

	user := &User{data: usr, acl: module}

	return user
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
			role.AddPermission(module.Permissions[p])
		}

		// Populate map with permissions
		module.Map.Add(role)
	}

	for name, rules := range module.Rules {
		if len(rules.Inherits) > 0 {
			module.Map.SetParents(name, rules.Inherits)
		}
	}

	return module
}
