package user

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"gopkg.in/mgo.v2/bson"
)

func Boot() *Module {

	module := &Module{}

	return module
}

type Module struct {
	Mongo  *mongo.Service `inject:""`
}

func (module *Module) Get(id bson.ObjectId) (*One, error) {

	var model *User
	context := module
	database := module.Mongo.Database

	// Get the user using it's id
	err := database.C("users").FindId(id).One(&model)

	if err != nil {
		
		return nil, exceptions.NotFound{"Invalid user id. Not found."}
	}

	user := &One{data: model, di: context}

	return user, nil
}
