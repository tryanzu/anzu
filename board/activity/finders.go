package activity

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

func Count(d deps, q bson.M) int {
	ctx := context.TODO()
	n, err := d.Mgo().Collection("activity").CountDocuments(ctx, q)
	if err != nil {
		panic(err)
	}
	return int(n)
}

func CountList(d deps, q bson.M) int {
	var result struct {
		Count int `bson:"count"`
	}
	q["list"] = bson.M{"$exists": true}
	ctx := context.TODO()
	cursor, err := d.Mgo().Collection("activity").Aggregate(ctx, []bson.M{
		{"$match": q},
		{"$project": bson.M{"size": bson.M{"$size": "$list"}}},
		{"$group": bson.M{"_id": "null", "count": bson.M{"$sum": "$size"}}},
	})
	if err != nil {
		log.Errorf("activity count	err=%v", err)
		return 0
	}
	defer cursor.Close(ctx)
	if cursor.Next(ctx) {
		err = cursor.Decode(&result)
		if err != nil {
			log.Errorf("activity count decode	err=%v", err)
			return 0
		}
	}
	return result.Count
}
