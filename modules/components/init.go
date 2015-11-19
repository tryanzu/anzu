package components

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/mongo"
)

func Boot() *Module {

	module := &Module{}

	return module
}

type Module struct {
	Mongo *mongo.Service `inject:""`
}

func (module Module) Get(find interface{}) (*Component, error) {

	var model *User
	context := module
	database := module.Mongo.Database

	switch usr.(type) {
	case bson.ObjectId:

		// Get the user using it's id
		err := database.C("users").FindId(usr.(bson.ObjectId)).One(&model)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid user id. Not found."}
		}

	case bson.M:

		// Get the user using it's id
		err := database.C("users").Find(usr.(bson.M)).One(&model)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid user id. Not found."}
		}

	case *User:

		model = usr.(*User)

	default:
		panic("Unkown argument")
	}

	user := &One{data: model, di: context}

	return user, nil
}
