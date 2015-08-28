package acl

import (
	"encoding/json"
	"github.com/mikespook/gorbac"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
)

type Module struct {
	Map   *gorbac.Rbac
	Mongo *mongo.Service     `inject:""`
	Rules map[string]AclRole
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

	for name, rules := range module.Rules {

		// Populate map with permissions
		module.Map.Add(name, rules.Permissions, rules.Inherits)
	}

	return module
}

