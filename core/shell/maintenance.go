package shell

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/tryanzu/core/board/comments"
	"github.com/tryanzu/core/board/votes"

	"github.com/abiosoft/ishell"
	"github.com/goware/emailx"
	"github.com/op/go-logging"
	coreUser "github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/user"
	"gopkg.in/mgo.v2/bson"
)

var (
	log = logging.MustGetLogger("shell")
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

func RunAnzuGarbageCollector(c *ishell.Context) {
	c.Print("running anzu garbage collector...")
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	db := deps.Container.Mgo()
	query := db.C("users").Find(bson.M{"validated": false, "scheduled_delete": bson.M{"$exists": false}})
	iter := query.Iter()
	var usr coreUser.User
	for iter.Next(&usr) {
		err := usr.EnforceAccountValidationEmail()
		if err != nil {
			c.Println("enforce account validation email failed", err)
			continue
		}
		err = db.C("users").UpdateId(usr.Id, bson.M{"$set": bson.M{"scheduled_delete": time.Now().Add(time.Hour * 24)}})
		if err != nil {
			c.Println("scheduled delete failed", err)
			continue
		}
	}
	query = db.C("users").Find(bson.M{"scheduled_delete": bson.M{"$lte": time.Now()}})
	count, _ := query.Count()
	iter = query.Iter()
	c.Printf("about to delete %v users", count)
	for iter.Next(&usr) {
		// lets move the user to another collection of deleted users.
		err := db.C("deleted_users").Insert(&usr)
		if err != nil {
			c.Printf("something went wrong deleting user %v: %v", usr.Id, err)
			continue
		}
		err = db.C("users").RemoveId(usr.Id)
		if err != nil {
			c.Printf("something went wrong deleting user %v: %v", usr.Id, err)
		}
		info, err := db.C("comments").UpdateAll(bson.M{"user_id": usr.Id}, bson.M{"deleted_at": time.Now()})
		if err != nil {
			c.Printf("something went wrong deleting comments %v: %v", usr.Id, err)
		}
		c.Printf("removed %v comments from %v", info.Updated, usr.Id)
		info, err = db.C("posts").UpdateAll(bson.M{"user_id": usr.Id}, bson.M{"deleted_at": time.Now()})
		if err != nil {
			c.Printf("something went wrong deleting comments %v: %v", usr.Id, err)
		}
		c.Printf("removed %v posts from %v", info.Updated, usr.Id)
		err = emailx.Validate(usr.Email)
		if err == nil {
			usr.UnvalidatedAccountDeletion()
		}
	}
}

func RebuildTrustNet(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)
	log.Info("rebuild trust net...")
	db := deps.Container.Mgo()
	cache := deps.Container.CacheProvider
	now := time.Now()
	query := db.C("users").Find(bson.M{"validated": true, "last_seen_at": bson.M{"$gte": time.Date(2020, 1, 1, 12, 0, 0, 0, now.Location())}}).Iter()
	var usr coreUser.User
	for query.Next(&usr) {
		usrVotes := db.C("votes").Find(bson.M{"user_id": usr.Id}).Iter()
		var vote votes.Vote
		for usrVotes.Next(&vote) {
			var c comments.Comment
			if vote.Type == "comment" {
				if vote.Value == "useful" || vote.Value == "goodExplanation" || vote.Value == "upvote" {
					err := db.C("comments").FindId(vote.RelatedID).One(&c)
					if err == nil {
						cmd := cache.XAdd(context.Background(), &redis.XAddArgs{
							Stream: "trustnet.assignment",
							ID:     "*",
							Values: []string{"src", usr.Id.Hex(), "dst", c.UserId.Hex(), "weight", "0.25"},
						})
						if cmd.Err() != nil {
							log.Error(cmd.Err())
						}
						log.Debugf("assigned xadd succeed 	usr=%s", usr.UserName)
					}
				}
			}
		}
	}
}
