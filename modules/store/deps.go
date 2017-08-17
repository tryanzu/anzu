package store

import (
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Deps interface {
	Mgo() *mgo.Database
	Mailer() mail.Mailer
}

type Query func(col *mgo.Collection) *mgo.Query
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

func WithinID(list []bson.ObjectId) Scope {
	return func(query bson.M) bson.M {
		query["_id"] = bson.M{"$in": list}
		return query
	}
}

func bsonq(scopes ...Scope) bson.M {
	query := bson.M{}

	// Apply all scopes to construct query.
	for _, s := range scopes {
		query = s(query)
	}
	return query
}
