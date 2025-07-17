package acl

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/mikespook/gorbac"
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/deps"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var LoadedACL *Module

type Module struct {
	Map         *gorbac.RBAC
	Rules       map[string]AclRole
	Permissions map[string]gorbac.Permission
}

func (module *Module) User(id primitive.ObjectID) *User {

	var usr model.User
	database := deps.Container.Mgo()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get the user using it's id
	err := database.Collection("users").FindOne(ctx, bson.M{"_id": id}).Decode(&usr)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
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
	rules, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	// Unmarshal file with gaming rules
	if err := json.Unmarshal(rules, &module.Rules); err != nil {
		panic(err)
	}

	module.Map = gorbac.New()
	module.Permissions = make(map[string]gorbac.Permission)

	for name, rules := range module.Rules {

		role := gorbac.NewStdRole(name)

		for _, p := range rules.Permissions {
			module.Permissions[p] = gorbac.NewStdPermission(p)
			_ = role.Assign(module.Permissions[p])
		}

		// Populate map with permissions
		_ = module.Map.Add(role)
	}

	for name, rules := range module.Rules {
		if len(rules.Inherits) > 0 {
			_ = module.Map.SetParents(name, rules.Inherits)
		}
	}

	LoadedACL = module
	return module
}
