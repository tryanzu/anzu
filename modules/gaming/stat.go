package gaming

import (
	"github.com/fernandez14/spartangeek-blacker/modules/user"
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

	database := self.Mongo.Database

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
	var rankings []RankingModel

	// Recover from any panic even inside this goroutine
	defer self.Errors.Recover()

	database := self.Mongo.Database
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

		rankings = append(rankings, RankingModel{
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
		})
	}

	sort.Sort(RankBySwords(rankings))

	p := 0
	
	for i := 0; i < len(rankings); i++ {
		p++
		rankings[i].Position.Swords = p

		log.Printf("[job] [ResetGeneralRanking] [Swords] User %v is %v \n", rankings[i].UserId.Hex(), p)
	}

	sort.Sort(RankByCoins(rankings))

	p = 0

	for i := 0; i < len(rankings); i++ {
		p++
		rankings[i].Position.Wealth = p
	}

	sort.Sort(RankByBadges(rankings))

	p = 0

	for i := 0; i < len(rankings); i++ {

		p++
		rankings[i].Position.Badges = p

		err := database.C("stats").Insert(rankings[i])

		if err != nil {
			panic(err)
		}
	}
}
