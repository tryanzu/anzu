package components

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/search"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"gopkg.in/mgo.v2/bson"
)

func Boot() *Module {

	module := &Module{}

	return module
}

type Module struct {
	Mongo  *mongo.Service `inject:""`
	Search *search.Module `inject:""`
}

func (module *Module) Get(find interface{}) (*ComponentModel, error) {

	var model interface{}

	context := module
	database := module.Mongo.Database

	switch find.(type) {
	case bson.ObjectId:

		// Get the user using it's id
		err := database.C("components").FindId(find.(bson.ObjectId)).One(&model)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid component id. Not found."}
		}

	case bson.M:

		// Get the user using it's id
		err := database.C("components").Find(find.(bson.M)).One(&model)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid component finder. Not found."}
		}

	default:
		panic("Unkown argument")
	}

	// Marshal the data inside the generic model
	encoded, err := bson.Marshal(model)

	if err != nil {
		panic(err)
	}

	var component *ComponentModel

	err = bson.Unmarshal(encoded, &component)

	if err != nil {
		panic(err)
	}

	component.SetDI(context)
	component.SetGeneric(encoded)

	return component, nil
}
