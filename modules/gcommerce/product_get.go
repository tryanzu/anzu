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

func (this Products) GetList(limit, offset int, search, category, kind string) ([]*Product, []ProductAggregation, int, error) {

	list := make([]*Product, 0)
	facets := make([]ProductAggregation, 0)

	query := bson.M{}
	fields := bson.M{}

	if kind != "" {
		query["type"] = kind
	}

	if category != "" {
		query["category"] = category
	}

	if search != "" {
		fields["score"] = bson.M{"$meta": "textScore"}
		query["$text"] = bson.M{"$search": search}
	}

	database := this.di.Mongo.Database

	if _, exists := query["$text"]; exists {

		err := database.C("gcommerce_products").Find(query).Select(fields).Sort("$textScore:score").Limit(limit).Skip(offset).All(&list)

		if err != nil {
			return list, facets, 0, err
		}

	} else {

		err := database.C("gcommerce_products").Find(query).Select(fields).Limit(limit).Skip(offset).All(&list)

		if err != nil {
			return list, facets, 0, err
		}
	}

	this.InitializeList(list)

	rows, err := database.C("gcommerce_products").Find(query).Count()

	if err != nil {
		return list, facets, 0, err
	}

	// Remove type from aggregation since its not needed
	delete(query, "type")
	delete(query, "category")

	err = database.C("gcommerce_products").Pipe([]bson.M{
		{"$match": query},
		{"$group": bson.M{"_id": "$category", "count": bson.M{"$sum": 1}}},
	}).All(&facets)

	if err != nil {
		return list, facets, 0, err
	}
	
	return list, facets, rows, nil
}
