package votes

import (
	"github.com/tryanzu/core/core/common"
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

// FindList of votes for given scopes.
func FindList(deps Deps, scopes ...common.Scope) (list List, err error) {
	err = deps.Mgo().C("votes").Find(common.ByScope(scopes...)).All(&list)
	return
}
