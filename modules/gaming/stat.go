package gaming

import (
	"context"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"sort"
	"time"
)

func (self *Module) GetRankingBy(sort string) []RankingModel {

	var rankings []RankingModel
	var users []RankingUserModel
	var users_id []primitive.ObjectID

	ctx := context.Background()
	database := deps.Container.Mgo()

	// Get the rankings with the sort parameter
	cursor, err := database.Collection("stats").Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"created_at": -1, "position." + sort: 1}).SetLimit(50))
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &rankings)
	if err != nil {
		panic(err)
	}

	for _, ranking := range rankings {
		users_id = append(users_id, ranking.UserId)
	}

	cursor, err = database.Collection("users").Find(ctx, bson.M{"_id": bson.M{"$in": users_id}}, options.Find().SetProjection(bson.M{"_id": 1, "username": 1, "image": 1, "gaming.level": 1}))
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &users)
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

	ctx := context.Background()
	database := deps.Container.Mgo()
	current_batch := time.Now()

	cursor, err := database.Collection("users").Find(ctx, bson.M{}, options.Find().SetBatchSize(1000))
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)

	log.Println("[job] [ResetGeneralRanking] Started")

	for cursor.Next(ctx) {
		err := cursor.Decode(&usr)
		if err != nil {
			panic(err)
		}

		log.Printf("[job] [ResetGeneralRanking] Processing user %v\n", usr.Id.Hex())

		var before RankingPositionModel
		var before_this RankingModel

		err = database.Collection("stats").FindOne(ctx, bson.M{"user_id": usr.Id}, options.FindOne().SetSort(bson.M{"created_at": -1})).Decode(&before_this)

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

		_, err := database.Collection("stats").InsertOne(ctx, rankings[item.Id])
		if err != nil {
			panic(err)
		}
	}
}
