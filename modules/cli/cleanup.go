package cli

import (
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/user"
	"gopkg.in/mgo.v2/bson"
)

func (m Module) CleanupReferences() {

	var usr user.UserPrivate

	db := deps.Container.Mgo()
	log := m.Logger
	log.Info("Starting to compute users with duplicated email")
	pipe := db.C("users").Find(bson.M{"duplicates": bson.M{"$exists": true}}).Iter()

	for pipe.Next(&usr) {
		if len(usr.Duplicates) > 0 {
			for _, dup := range usr.Duplicates {
				changed, err := db.C("posts").UpdateAll(bson.M{"user_id": dup}, bson.M{"$set": bson.M{"user_id": usr.Id}, "$pull": bson.M{"users": dup}})

				if err != nil {
					log.Critical("Error while updating posts: %v", err)
					continue
				}

				log.Debug("[%s][posts] Ref %s updated %v", usr.UserName, dup.Hex(), changed.Updated)

				changed, err = db.C("comments").UpdateAll(bson.M{"user_id": dup}, bson.M{"$set": bson.M{"user_id": usr.Id}})

				if err != nil {
					log.Critical("Error while updating comments: %v", err)
					continue
				}

				log.Debug("[%s][comments] Ref %s updated %v", usr.UserName, dup.Hex(), changed.Updated)

				changed, err = db.C("mentions").UpdateAll(bson.M{"user_id": dup}, bson.M{"$set": bson.M{"user_id": usr.Id}})

				if err != nil {
					log.Critical("Error while updating mentions-ow: %v", err)
					continue
				}

				log.Debug("[%s][mentions-ow] Ref %s updated %v", usr.UserName, dup.Hex(), changed.Updated)

				changed, err = db.C("mentions").UpdateAll(bson.M{"from_id": dup}, bson.M{"$set": bson.M{"from_id": usr.Id}})

				if err != nil {
					log.Critical("Error while updating mentions: %v", err)
					continue
				}

				log.Debug("[%s][mentions] Ref %s updated %v", usr.UserName, dup.Hex(), changed.Updated)
			}

			continue
		}

		log.Info("User %s (%s) has no duplicates", usr.UserName, usr.Id.Hex())
	}
}
