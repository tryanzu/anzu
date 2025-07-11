package votes

import (
	"context"
	"github.com/tryanzu/core/core/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// FindVotableByUser gets the votes for ref.
func FindVotableByUser(deps Deps, votable Votable, userID primitive.ObjectID) (vote Vote, err error) {
	ctx := context.TODO()
	err = deps.Mgo().Collection("votes").FindOne(ctx, bson.M{
		"type":       votable.VotableType(),
		"related_id": votable.VotableID(),
		"user_id":    userID,
	}).Decode(&vote)
	if err == mongo.ErrNoDocuments {
		err = nil // No vote found is not an error
	}
	return
}

// FindList of votes for given scopes.
func FindList(deps Deps, scopes ...common.Scope) (list List, err error) {
	ctx := context.TODO()
	cursor, err := deps.Mgo().Collection("votes").Find(ctx, common.ByScope(scopes...))
	if err != nil {
		return
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &list)
	return
}
