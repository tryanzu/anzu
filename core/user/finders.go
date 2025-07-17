package user

import (
	"context"
	"errors"
	"time"

	"github.com/tryanzu/core/core/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var UserNotFound = errors.New("User has not been found by given criteria.")

func FindId(d deps, id primitive.ObjectID) (user User, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = d.Mgo().Collection("users").FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return user, UserNotFound
		}
		return user, err
	}

	return
}

func FindEmail(d deps, email string) (user User, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = d.Mgo().Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return user, UserNotFound
		}
		return user, err
	}

	return
}

func FindList(d deps, scopes ...common.Scope) (users Users, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := d.Mgo().Collection("users").Find(ctx, common.ByScope(scopes...))
	if err != nil {
		return users, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &users)
	return
}

func FetchBy(d deps, query common.Query) (UsersSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Count documents
	c, err := d.Mgo().Collection("users").CountDocuments(ctx, bson.M{})
	if err != nil {
		return UsersSet{}, err
	}

	// Find documents
	cursor, err := query(d.Mgo().Collection("users"), ctx)
	if err != nil {
		return UsersSet{}, err
	}
	defer cursor.Close(ctx)

	list := Users{}
	err = cursor.All(ctx, &list)
	if err != nil {
		return UsersSet{}, err
	}
	return UsersSet{
		Count: int(c),
		List:  list,
	}, nil
}

func Page(limit int, reverse bool, before *primitive.ObjectID, after *primitive.ObjectID) common.Query {
	return func(col *mongo.Collection, ctx context.Context) (*mongo.Cursor, error) {
		criteria := bson.M{
			"deleted_at": bson.M{"$exists": false},
		}

		if before != nil {
			criteria["_id"] = bson.M{"$lt": *before}
		}

		if after != nil {
			criteria["_id"] = bson.M{"$gt": *after}
		}

		findOptions := options.Find().SetLimit(int64(limit)).SetSkip(0)

		if reverse {
			findOptions.SetSort(bson.M{"created_at": -1})
		} else {
			findOptions.SetSort(bson.M{"created_at": 1})
		}

		return col.Find(ctx, criteria, findOptions)
	}
}

func FindNames(d deps, list ...primitive.ObjectID) (common.UsersStringMap, error) {
	hash := common.UsersStringMap{}
	missing := []primitive.ObjectID{}

	for _, id := range list {
		v, err := d.LedisDB().Get([]byte("user:" + id.Hex() + ":names"))
		if err == nil && len(v) > 0 {
			hash[id] = string(v)
			continue
		}

		// Append to list of missing keys
		missing = append(missing, id)
	}

	if len(missing) == 0 {
		return hash, nil
	}

	users, err := FindList(d, common.WithinID(missing))
	if err != nil {
		return hash, err
	}

	err = users.UpdateCache(d)
	if err != nil {
		return hash, err
	}

	for _, u := range users {
		hash[u.Id] = u.UserName
	}

	// Unknown users should be cached like so...
	if len(missing) != len(users) {
		for _, id := range missing {
			if _, exists := hash[id]; !exists {
				hash[id] = "Unknown"
				err = d.LedisDB().Set([]byte("user:"+id.Hex()+":names"), []byte("Unknown"))
				if err != nil {
					return hash, err
				}
			}
		}
	}

	return hash, nil
}
