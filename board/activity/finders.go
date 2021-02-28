package activity

import (
	"gopkg.in/mgo.v2/bson"
)

func Count(d deps, q bson.M) int {
	n, err := d.Mgo().C("activity").Find(q).Count()
	if err != nil {
		panic(err)
	}
	return n
}

func CountList(d deps, q bson.M) int {
	var result struct {
		Count int `bson:"count"`
	}
	q["list"] = bson.M{"$exists": true}
	err := d.Mgo().C("activity").Pipe([]bson.M{
		{"$match": q},
		{"$project": bson.M{"size": bson.M{"$size": "$list"}}},
		{"$group": bson.M{"_id": "null", "count": bson.M{"$sum": "$size"}}},
	}).One(&result)
	if err != nil {
		log.Errorf("activity count	err=%v", err)
	}
	return result.Count
}
