package users

import (
	"html"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// UpsertBan performs validations before upserting data struct
func UpsertBan(d deps, ban Ban) (Ban, error) {
	if ban.ID.Valid() == false {
		ban.ID = bson.NewObjectId()
		ban.Created = time.Now()
		ban.Status = ACTIVE
	}

	ban.Content = html.EscapeString(ban.Content)
	ban.Updated = time.Now()
	changes, err := d.Mgo().C("bans").UpsertId(ban.ID, bson.M{"$set": ban})
	if err != nil {
		return ban, err
	}

	if changes.Matched == 0 {
		// When inserted
	}
	return ban, nil
}
