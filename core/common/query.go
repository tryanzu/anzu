package common

import (
	"context"
	
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Query func(col *mongo.Collection, ctx context.Context) (*mongo.Cursor, error)
type Scope func(bson.M) bson.M

func SoftDelete(query bson.M) bson.M {
	query["deleted_at"] = bson.M{"$exists": false}
	return query
}

func FulltextSearch(search string) Scope {
	return func(query bson.M) bson.M {
		query["$text"] = bson.M{"$search": search}
		return query
	}
}

func FieldExists(field string, exists bool) Scope {
	return func(query bson.M) bson.M {
		query[field] = bson.M{"$exists": exists}
		return query
	}
}

func WithinID(list []primitive.ObjectID) Scope {
	return func(query bson.M) bson.M {
		query["_id"] = bson.M{"$in": list}
		return query
	}
}

func ByScope(scopes ...Scope) bson.M {
	query := bson.M{}

	// Apply all scopes to construct query.
	for _, s := range scopes {
		query = s(query)
	}
	return query
}
