package votes

import (
	"gopkg.in/mgo.v2/bson"
)

// FindVotableByUser gets the votes for ref.
func FindVotableByUser(deps Deps, votable Votable, userID bson.ObjectId) (vote Vote, err error) {
	err = deps.Mgo().C("votes").Find(bson.M{
		"type":       votable.VotableType(),
		"related_id": votable.VotableID(),
		"user_id":    userID,
	}).One(&vote)
	return
}
