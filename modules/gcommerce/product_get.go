package gcommerce

import (
	"gopkg.in/mgo.v2/bson"

	"time"
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

func (this Products) GetByBson(query bson.M) (*Product, error) {

	var model *Product

	database := this.di.Mongo.Database
	err := database.C("gcommerce_products").Find(query).One(&model)

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

	query := bson.M{"stock": bson.M{"$ne": 0}}
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

func (this Products) GetMassdrops(limit, offset int) []MassdropFoundation {

	var list []MassdropFoundation
	var ids []bson.ObjectId
	var prods []bson.ObjectId

	database := this.di.Mongo.Database
	err := database.C("gcommerce_massdrop").Find(nil).Limit(limit).Skip(offset).All(&list)

	if err != nil {
		panic(err)
	}

	for _, m := range list {
		ids = append(ids, m.Id)
		prods = append(prods, m.ProductId)
	}

	var aggregation []MassdropAggregation
	var insterested_map map[string]int = make(map[string]int)
	var reservation_map map[string]int = make(map[string]int)

	err = database.C("gcommerce_massdrop_transactions").Pipe([]bson.M{
		{"$match": bson.M{"massdrop_id": bson.M{"$in": ids}, "status": "completed"}},
		{"$group": bson.M{"_id": bson.M{"massdrop_id": "$massdrop_id", "type": "$type"}, "count": bson.M{"$sum": 1}}},
	}).All(&aggregation)

	if err != nil {
		panic(err)
	}

	for _, a := range aggregation {
		id := a.Id.MassdropID.Hex()

		if a.Id.Type == "interested" {
			insterested_map[id] = a.Count
		} else if a.Id.Type == "reservation" {
			reservation_map[id] = a.Count
		}
	}

	products_map := make(map[string]*Product)
	products := make([]*Product, 0)
	err = database.C("gcommerce_products").Find(bson.M{"_id": bson.M{"$in": prods}}).All(&products)

	if err != nil {
		panic(err)
	}

	this.InitializeList(products)

	for _, p := range products {
		products_map[p.Id.Hex()] = p
	}

	for index, m := range list {

		mp := products_map[m.ProductId.Hex()]
		interested, ie := insterested_map[m.Id.Hex()]

		if !ie {
			interested = 0
		}

		reservations, re := reservation_map[m.Id.Hex()]

		if !re {
			reservations = 0
		}

		list[index].Name = mp.Name
		list[index].Slug = mp.Slug
		list[index].Reservations = reservations
		list[index].Interested = interested

		// Keep starting price
		list[index].StartingPrice = m.Price

		for ci, checkpoint := range m.Checkpoints {

			if reservations >= checkpoint.Starts {
				list[index].Checkpoints[ci].Done = true
				list[index].Price = checkpoint.Price
				list[index].Deadline = list[index].Deadline.Add(time.Duration(checkpoint.Timespan) * time.Hour)
			}
		}

		// Deactivate massdrop when deadline has been reached
		if list[index].Deadline.Before(time.Now()) {
			list[index].Active = false
		}
	}

	return list
}
