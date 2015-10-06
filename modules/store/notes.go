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

func (module *Module) CreateNote(title, content string) error {

	note := &BuildResponseModel{
		Title: title,
		Content: content,
	}

	database := module.Mongo.Database
	err := database.C("builds_responses").Insert(note)

	return err
}

func (module *Module) UpdateNote(id bson.ObjectId, title, content string) error {

	database := module.Mongo.Database
	err := database.C("builds_responses").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"title": title, "content": content}})

	return err
}
