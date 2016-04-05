package products

import (
	"gopkg.in/mgo.v2/bson"
)

func (this Module) GetById(id bson.ObjectId) (*Product, error) {

	var model *Product

	database := this.di.Mongo.Database
	err := database.C("gcommerce_products").Find(bson.M{"_id": id}).One(&model)

	if err != nil {
		return nil, err
	}

	model.SetDI(this.di)
	
	return model, nil
}
