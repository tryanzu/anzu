package votes

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// VoteType should be an integer in the form of up or down.
type VoteType string

func isValidVoteType(str string) bool {
	for _, t := range []string{"useful", "offtopic", "wordy", "concise"} {
		if t == str {
			return true
		}
	}
	return false
}

type voteStatus struct {
	Count  int  `json:"count"`
	Active bool `json:"active"`
}

// UpsertVote creates or removes a vote for given votable item<->user
func UpsertVote(deps Deps, item Votable, userID bson.ObjectId, kind string) (vote Vote, status voteStatus, err error) {
	if isValidVoteType(kind) == false {
		err = errors.New("invalid vote type")
		return
	}

	criteria := bson.M{
		"type":       item.VotableType(),
		"related_id": item.VotableID(),
		"value":      kind,
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

	status = voteStatus{
		Active: true,
		Count:  0,
	}

	// Get current vote status from remote.
	err = coll(deps).Find(criteria).One(&vote)
	if err != nil {
		panic(err)
	}
	delete(criteria, "user_id")
	criteria["deleted_at"] = bson.M{"$exists": false}
	c, err := coll(deps).Find(criteria).Count()
	if err != nil {
		panic(err)
	}
	status.Count = c

	// Delete when the vote is not new. (toggle)
	if changes.Matched > 0 && vote.Deleted == nil {
		deleted := time.Now()
		vote.Deleted = &deleted
		status.Active = false
		status.Count--

		err = coll(deps).UpdateId(vote.ID, bson.M{"$set": bson.M{"deleted_at": deleted}})
		if err != nil {
			panic(err)
		}
		return
	}
	if vote.Deleted != nil {
		status.Count++
	}
	err = coll(deps).UpdateId(vote.ID, bson.M{"$unset": bson.M{"deleted_at": 1}})
	if err != nil {
		panic(err)
	}
	vote.Deleted = nil
	return
}
