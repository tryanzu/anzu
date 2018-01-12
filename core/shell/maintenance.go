package shell

import (
	"strconv"
	"strings"

	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/user"
	"gopkg.in/abiosoft/ishell.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	CMD_NONE   = "CMD_NONE"
	CMD_SKIP   = "CMD_SKIP"
	CMD_CANCEL = "CMD_CANCEL"
	CMD_UPDATE = "CMD_UPDATE"
)

func CleanupDuplicatedEmails(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	db := deps.Container.Mgo()
	pipe := db.C("users").Pipe([]bson.M{
		{
			"$match": bson.M{
				"email":      bson.M{"$exists": true, "$ne": ""},
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
				"count": bson.M{"$gte": 2},
			},
		},
	}).Iter()

	var duplicate struct {
		Email string          `bson:"_id"`
		List  []bson.ObjectId `bson:"uniqueIds"`
		Count int             `bson:"count"`
	}

	for pipe.Next(&duplicate) {
		var users []user.UserPrivate

		err := db.C("users").Find(bson.M{"_id": bson.M{"$in": duplicate.List}}).Sort("-created_at").All(&users)
		if err != nil {
			c.Println("Could not get list", err)
			continue
		}

		if len(users) < 2 {
			c.Println("Could not get list of more of 2 users, skipping...")
			continue
		}

		c.Println("Disputed email " + duplicate.Email + " has " + strconv.Itoa(len(duplicate.List)) + " duplicates:")
		for index, usr := range users {
			c.Printf("%s) %s::%s (%s) \n Created: %s\n Validated: %v\n Swords: %v\n Level: %v\n--------\n",
				strconv.Itoa(index), usr.Id.Hex(),
				usr.UserName,
				usr.UserNameSlug,
				usr.Created,
				usr.Validated,
				usr.Gaming.Swords,
				usr.Gaming.Level,
			)
		}

		action := CMD_NONE
		choosed := 0

		// For until a valid command is ready to execute.
		for action == CMD_NONE {
			c.Print("Who's the winner?  S(kip) C(ancel) or $index: ")
			option := strings.ToLower(c.ReadLine())
			if option == "skip" || option == "s" {
				action = CMD_SKIP
				break
			}

			if option == "cancel" || option == "c" {
				action = CMD_CANCEL
				break
			}

			index, err := strconv.Atoi(option)
			if err != nil {
				c.Println("Invalid option, S(kip) C(ancel) or $index...")
				continue
			}

			if len(users) < index {
				c.Println("Option not found, S(kip) C(ancel) or $index...")
				continue
			}

			choosed = index
			action = CMD_UPDATE
		}

		if action == CMD_SKIP {
			continue
		}

		if action == CMD_CANCEL {
			break
		}

		id := users[choosed].Id
		for _, usr := range users {
			if id == usr.Id {
				continue
			}

			_, err := db.C("users_deleted_list").UpsertId(usr.Id, usr)
			if err != nil {
				c.Println("Could not backup user", err)
				continue
			}

			err = db.C("users_deleted_list").UpdateId(usr.Id, bson.M{"$set": bson.M{"kept_user_id": id}})
			if err != nil {
				c.Println("Could not backup user", err)
				continue
			}

			err = db.C("users").RemoveId(usr.Id)
			if err != nil {
				c.Println("Could not remove user", err)
				continue
			}
		}
	}
}
