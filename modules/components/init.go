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

func (module *Module) SearchComponents(content string) ([]Component, []ComponentTypeCountModel) {

	components := make([]Component, 0)
	count := make([]ComponentTypeCountModel, 0)
	database := module.Mongo.Database

	// Fields to retrieve
	fields := ComponentFields
	fields["score"] = bson.M{"$meta": "textScore"}

	query := bson.M{
		"$text": bson.M{"$search": content},
	}

	err := database.C("components").Find(query).Select(fields).Sort("$textScore:score").Limit(10).All(&components)

	if err != nil {
		panic(err)
	}

	err = database.C("components").Pipe([]bson.M{
		{"$match": query},
		{"$sort": bson.M{"score": bson.M{"$meta": "textScore"}}},
		{"$group": bson.M{"_id": "$type", "count": bson.M{"$sum": 1}}},
	}).All(&count)

	if err != nil {
		panic(err)
	}

	return components, count
}