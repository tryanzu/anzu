package gcommerce

import (
	"gopkg.in/mgo.v2/bson"
)

func (this Products) GetById(id bson.ObjectId) (*Product, error) {

	var model *Product

	database := this.di.Mongo.Database
	err := database.C("gcommerce_products").Find(bson.M{"_id": id}).One(&model)

	if err != nil {
		return nil, err
	}

	model.SetDI(this.di)
	model.Initialize()
	
	return model, nil
}
