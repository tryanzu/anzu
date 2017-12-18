package gaming

import (
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/user"
	"gopkg.in/mgo.v2/bson"
	"log"
	"sort"
	"time"
)

func (self *Module) GetRankingBy(sort string) []RankingModel {

	var ranking RankingModel
	var rankings []RankingModel
	var users []RankingUserModel
	var users_id []bson.ObjectId

	database := deps.Container.Mgo()

	// Get the rankings with the sort parameter
	iter := database.C("stats").Find(nil).Sort("-created_at", "position."+sort).Limit(50).Iter()

	for iter.Next(&ranking) {

		rankings = append(rankings, ranking)
		users_id = append(users_id, ranking.UserId)
	}

	err := database.C("users").Find(bson.M{"_id": bson.M{"$in": users_id}}).Select(bson.M{"_id": 1, "username": 1, "image": 1, "gaming.level": 1}).All(&users)

	if err != nil {
		panic(err)
	}

	for id, rank := range rankings {

		for _, user := range users {

			if user.Id == rank.UserId {

				rankings[id].User = user

				break
			}
		}
	}

	return rankings
}

func (self *Module) ResetGeneralRanking() {

	var usr user.User
	var swordsRank RankPositions
	var wealthRank RankPositions
	var badgesRank RankPositions

	rankings := map[string]RankingModel{}

	// Recover from any panic even inside this goroutine
	defer self.Errors.Recover()

	database := deps.Container.Mgo()
	current_batch := time.Now()

	iter := database.C("users").Find(nil).Batch(1000).Prefetch(0.50).Iter()

	log.Println("[job] [ResetGeneralRanking] Started")

	for iter.Next(&usr) {

		log.Printf("[job] [ResetGeneralRanking] Processing user %v\n", usr.Id.Hex())

		var before RankingPositionModel
		var before_this RankingModel

		err := database.C("stats").Find(bson.M{"user_id": usr.Id}).Sort("-created_at").Limit(1).One(&before_this)

		if err != nil {

			before = RankingPositionModel{
				Wealth: 0,
				Badges: 0,
				Swords: 0,
			}

		} else {

			before = before_this.Position
		}

		rankings[usr.Id.Hex()] = RankingModel{
			UserId: usr.Id,
			Badges: len(usr.Gaming.Badges),
			Swords: usr.Gaming.Swords,
			Coins:  usr.Gaming.Coins,
			Position: RankingPositionModel{
				Wealth: 0,
				Badges: 0,
				Swords: 0,
			},
			Before:  before,
			Created: current_batch,
		}

		swordsRank = append(swordsRank, RankPosition{
			Id:    usr.Id.Hex(),
			Value: usr.Gaming.Swords,
		})

		wealthRank = append(wealthRank, RankPosition{
			Id:    usr.Id.Hex(),
			Value: usr.Gaming.Coins,
		})

		badgesRank = append(badgesRank, RankPosition{
			Id:    usr.Id.Hex(),
			Value: len(usr.Gaming.Badges),
		})
	}

	sort.Sort(swordsRank)
	sort.Sort(wealthRank)
	sort.Sort(badgesRank)

	for pos, item := range swordsRank {

		r := rankings[item.Id]
		r.Position.Swords = pos + 1
		rankings[item.Id] = r

		log.Printf("[job] [ResetGeneralRanking] [Swords] User %v is %v \n", item.Id, pos)
	}

	for pos, item := range wealthRank {

		r := rankings[item.Id]
		r.Position.Wealth = pos + 1
		rankings[item.Id] = r

		log.Printf("[job] [ResetGeneralRanking] [Wealth] User %v is %v \n", item.Id, pos)
	}

	for pos, item := range badgesRank {

		r := rankings[item.Id]
		r.Position.Badges = pos + 1
		rankings[item.Id] = r

		log.Printf("[job] [ResetGeneralRanking] [Badges] User %v is %v \n", item.Id, pos)

		err := database.C("stats").Insert(rankings[item.Id])
		if err != nil {
			panic(err)
		}
	}
}
