package votes

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// VoteType should be an integer in the form of up or down.
type VoteType int

// UpsertVote creates or removes a vote for given votable item<->user
func UpsertVote(deps Deps, item Votable, userID bson.ObjectId, kind VoteType) (vote Vote, err error) {
	criteria := bson.M{
		"type":       item.VotableType(),
		"related_id": item.VotableID(),
		"user_id":    userID,
	}

	changes, err := coll(deps).Upsert(criteria, bson.M{
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
	})

	if err != nil {
		return
	}

	// Get current vote status from remote.
	err = coll(deps).Find(criteria).One(&vote)
	if err != nil {
		panic(err)
	}

	// Delete when the vote is not new. (toggle)
	if changes.Matched > 0 && vote.Deleted == nil {
		deleted := time.Now()
		vote.Deleted = &deleted

		err = coll(deps).UpdateId(vote.ID, bson.M{"$set": bson.M{"deleted_at": deleted}})
		if err != nil {
			panic(err)
		}

		return
	}

	return
}
