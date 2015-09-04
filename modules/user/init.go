package user

import (
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"gopkg.in/mgo.v2/bson"
)

func Boot() *Module {

	module := &Module{}

	return module
}

type Module struct {
	Mongo *mongo.Service `inject:""`
}

func (module *Module) Get(id bson.ObjectId) *One {

	var model *User
	database := module.Mongo.Database

	// Get the user using it's id
	err := database.C("users").FindId(id).One(&model)

	if err != nil {
		panic(err)
	}

	user := &One{data: model}

	return user
}
