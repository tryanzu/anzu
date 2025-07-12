package gaming

import (
	"context"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/exceptions"
	"github.com/tryanzu/core/modules/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func (self *User) AcquireBadge(id primitive.ObjectID, validation bool) error {

	var badge BadgeModel

	ctx := context.Background()
	database := deps.Container.Mgo()
	usr := self.user.Data()

	// Find the badge using it's id
	err := database.Collection("badges").FindOne(ctx, bson.M{"_id": id}).Decode(&badge)

	if err != nil {
		return exceptions.NotFound{"Invalid badge id, not found."}
	}

	if validation {

		if badge.Type != "clothes" && badge.Type != "weapon" && badge.Type != "power" && badge.Type != "armour" {

			return exceptions.UnexpectedValue{"Not a valid type of badge to get acquired."}
		}

		if badge.Coins > 0 && usr.Gaming.Coins < badge.Coins {

			return exceptions.OutOfBounds{"Not enough coins to buy item."}
		}

		if badge.RequiredLevel > 0 && usr.Gaming.Level < badge.RequiredLevel {

			return exceptions.OutOfBounds{"Not enough level."}
		}

		if !badge.RequiredBadge.IsZero() {

			var user_valid bool = false

			user_badges := usr.Gaming.Badges

			for _, user_badge := range user_badges {

				if user_badge.Id == badge.RequiredBadge {

					user_valid = true
				}
			}

			if !user_valid {

				return exceptions.OutOfBounds{"Don't have required badge."}
			}
		}
	}

	badge_push := user.UserBadge{
		Id:   id,
		Date: time.Now(),
	}

	_, err = database.Collection("users").UpdateOne(ctx, bson.M{"_id": usr.Id}, bson.M{"$push": bson.M{"gaming.badges": badge_push}})

	if err != nil {
		panic(err)
	}

	if badge.Coins > 0 {

		// Pay the badge price using user coins
		go self.Coins(-badge.Coins)
	}

	return nil
}
