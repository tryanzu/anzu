package cli

import (
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"gopkg.in/mgo.v2/bson"

	"time"
)

func (m Module) CleanupDuplicates() {

	var duplicate struct {
		Email string          `bson:"email"`
		List  []bson.ObjectId `bson:"uniqueIds"`
		Count int             `bson:"count"`
	}

	db := m.Mongo.Database
	log := m.Logger
	log.Info("Starting to compute users with duplicated email")
	pipe := db.C("users").Pipe([]bson.M{
		{
			"$match": bson.M{
				"email":      bson.M{"$ne": "", "$exists": true},
				"deleted_at": bson.M{"$exists": false},
			},
		},
		{
			"$group": bson.M{
				"_id":       "$email",
				"uniqueIds": bson.M{"$addToSet": "$_id"},
				"count":     bson.M{"$sum": 1},
			},
		},
		{
			"$match": bson.M{
				"count": bson.M{"$gt": 1},
			},
		},
	}).Iter()

	for pipe.Next(&duplicate) {
		var users []user.UserPrivate

		err := db.C("users").Find(bson.M{"_id": bson.M{"$in": duplicate.List}}).All(&users)
		if err != nil {
			log.Critical("Something wen't wrong: %v\n", err)
			continue
		}

		var elected *user.UserPrivate
		var gaming user.UserGaming

		fbReady := false
		username := ""

		if len(users) < 2 {
			log.Critical("Duplicate did not found two users to work with \n")
			continue
		}

		// Iterate over users to elect a base user which will be updated
		for _, usr := range users {
			log.Debug("[candidate] %s (%s) %+v %v\n", usr.UserName, usr.Email, usr.Gaming, usr.Facebook)

			switch usr.Facebook.(type) {
			case interface{}:
				data := usr.Facebook.(bson.M)
				fbUsername := helpers.StrSlug(data["first_name"].(string) + " " + data["last_name"].(string))
				if username == "" || fbUsername != usr.UserName {
					username = usr.UserName
				}

				if elected == nil && fbReady == false {
					elected = &usr
					fbReady = true
				} else if elected != nil && fbReady == true {
					log.Info("Skipping facebook duplicated information from %s, given: %+v\n", usr.UserName, data)
				}
			case nil:
				username = usr.UserName
				if elected == nil {
					elected = &usr
				}
			}

			if elected != nil && elected.Id != usr.Id && elected.Gaming.Swords < usr.Gaming.Swords && elected.Facebook == nil {
				elected = &usr
			}

			if usr.Gaming.Swords >= gaming.Swords {
				gaming = usr.Gaming
			}
		}

		if elected == nil {
			log.Critical("Election had no results %+v\n", duplicate.List)
			continue
		}

		dup := duplicate.List[:0]
		for _, i := range duplicate.List {
			if i != elected.Id {
				dup = append(dup, i)
			}
		}

		for _, usr := range users {
			if usr.Id == elected.Id {
				err := db.C("users").Update(bson.M{"_id": usr.Id}, bson.M{"$set": bson.M{"username": username, "gaming": gaming, "duplicates": dup}})
				if err != nil {
					panic(err)
				}

				continue
			}

			err := db.C("users").Update(bson.M{"_id": usr.Id}, bson.M{"$unset": bson.M{"email": "", "username": ""}, "$set": bson.M{"deleted_at": time.Now(), "duplicate.elected": elected.Id, "duplicate.email": usr.Email, "duplicate.username": usr.UserName, "duplicate.gaming": usr.Gaming}})
			if err != nil {
				panic(err)
			}
		}
	}
}
