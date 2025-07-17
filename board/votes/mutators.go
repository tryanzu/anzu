package votes

import (
	"context"
	"errors"
	"time"

	"github.com/tryanzu/core/core/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// VoteType should be an integer in the form of up or down.
type VoteType string

func isValidVoteType(str string) bool {
	cnf := config.C.Copy()
	return cnf.Site.IsValidReaction(str)
}

type voteStatus struct {
	Count  int  `json:"count"`
	Active bool `json:"active"`
}

// UpsertVote creates or removes a vote for given votable item<->user
func UpsertVote(deps Deps, item Votable, userID primitive.ObjectID, kind string) (vote Vote, status voteStatus, err error) {
	if !isValidVoteType(kind) {
		err = errors.New("invalid vote type")
		return
	}

	ctx := context.TODO()
	criteria := bson.M{
		"type":       item.VotableType(),
		"related_id": item.VotableID(),
		"value":      kind,
		"user_id":    userID,
	}

	update := bson.M{
		"$inc": bson.M{"changes": 1},
		"$set": bson.M{
			"type":       item.VotableType(),
			"related_id": item.VotableID(),
			"user_id":    userID,
			"value":      kind,
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}

	upsertTrue := true
	result, err := coll(deps).UpdateOne(ctx, criteria, update, &options.UpdateOptions{Upsert: &upsertTrue})
	if err != nil {
		return
	}

	status = voteStatus{
		Active: true,
		Count:  0,
	}

	// Get current vote status from remote.
	err = coll(deps).FindOne(ctx, criteria).Decode(&vote)
	if err != nil {
		panic(err)
	}
	delete(criteria, "user_id")
	criteria["deleted_at"] = bson.M{"$exists": false}
	c, err := coll(deps).CountDocuments(ctx, criteria)
	if err != nil {
		panic(err)
	}
	status.Count = int(c)

	// Delete when the vote is not new. (toggle)
	if result.MatchedCount > 0 && vote.Deleted == nil {
		deleted := time.Now()
		vote.Deleted = &deleted
		status.Active = false
		status.Count--

		_, err = coll(deps).UpdateOne(ctx, bson.M{"_id": vote.ID}, bson.M{"$set": bson.M{"deleted_at": deleted}})
		if err != nil {
			panic(err)
		}
		return
	}
	if vote.Deleted != nil {
		status.Count++
	}
	_, err = coll(deps).UpdateOne(ctx, bson.M{"_id": vote.ID}, bson.M{"$unset": bson.M{"deleted_at": 1}})
	if err != nil {
		panic(err)
	}
	vote.Deleted = nil
	return
}
