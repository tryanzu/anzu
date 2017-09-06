package user

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

func CanBeTrusted(user User) bool {
	return user.Warnings < 6
}

func TrustUserPosts(deps Deps, user User) {
	yesterday := time.Now().Add(-24 * time.Hour)

	var messyPost struct {
		Count float64 `bson:"count"`
	}

	// Count the number of post appearing to be spam or not legit.
	err := deps.Mgo().C("posts").Pipe([]bson.M{
		{
			"$sort": bson.M{"created_at": -1},
		},
		{
			"$match": bson.M{
				"created_at": bson.M{"$gte": yesterday},
				"user_id":    user.Id,
				"$or":        []bson.M{{"comments.count": 0, "users": bson.M{"$size": 1}}},
			},
		},
	}).One(&messyPost)

	if err != nil {
		panic(err)
	}

}
