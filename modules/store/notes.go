package store

import (
	"gopkg.in/mgo.v2/bson"
)

func (module *Module) GetNotes() []BuildResponseModel {

	var list []BuildResponseModel

	database := module.Mongo.Database
	err := database.C("builds_responses").Find(bson.M{}).Sort("-updated_at").All(&list)

	if err != nil {
		panic(err)
	}

	return list
}

func (module *Module) GetNote(id bson.ObjectId) (*BuildResponseModel, error) {

	var one *BuildResponseModel

	database := module.Mongo.Database
	err := database.C("builds_responses").FindId(id).One(&one)

	if err != nil {

		return nil, err
	}

	return one, nil
}

func (module *Module) CreateNote(title, content string, price int) error {

	note := &BuildResponseModel{
		Title: title,
		Content: content,
		Price: price,
	}

	database := module.Mongo.Database
	err := database.C("builds_responses").Insert(note)

	return err
}

func (module *Module) UpdateNote(id bson.ObjectId, title, content string, price int) error {

	database := module.Mongo.Database
	err := database.C("builds_responses").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"title": title, "content": content, "price": price}})

	return err
}

func (module *Module) DeleteNote(id bson.ObjectId) error {

	database := module.Mongo.Database
	err := database.C("builds_responses").RemoveId(id)

	return err
}
