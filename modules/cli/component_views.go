package cli

import (
	"gopkg.in/mgo.v2/bson"
)

func (module Module) GenerateComponentViews() {

	var counter struct {
		Id    bson.ObjectId `bson:"_id" json:"id"`
		Count int           `bson:"count" json:"count"`
	}

	database := module.Mongo.Database
	pipeline := database.C("user_views").Pipe([]bson.M{
		{
			"$match": bson.M{
				"related": "component",
			},
		},
		{
			"$group": bson.M{
				"_id":   "$related_id",
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{
				"count": -1,
			},
		},
	}).Iter()

	for pipeline.Next(&counter) {

		err := database.C("components").Update(bson.M{"_id": counter.Id}, bson.M{"$set": bson.M{"views": counter.Count}})

		if err != nil {
			panic(err)
		}

		module.Logger.Debugf("Moved counter of %s\n", counter.Id.Hex())
	}
}
