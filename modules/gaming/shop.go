package gaming

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func (self *User) AcquireBadge(id bson.ObjectId) error {

	var badge BadgeModel

	database := self.di.Mongo.Database
	usr := self.user.Data()

	// Find the badge using it's id
	err := database.C("badges").Find(bson.M{"_id": id}).One(&badge)

	if err != nil {

		return exceptions.NotFound{"Invalid badge id, not found."}
	}

	if badge.Type != "clothes" && badge.Type != "weapon" && badge.Type != "shield" {

		return exceptions.UnexpectedValue{"Not a valid type of badge to get acquired."}
	}

	if badge.Coins > 0 && usr.Gaming.Coins < badge.Coins {

		return exceptions.OutOfBounds{"Not enough coins to buy item."}
	}

	if badge.RequiredLevel > 0 && usr.Gaming.Level < badge.RequiredLevel {
		
		return exceptions.OutOfBounds{"Not enough level."}
	}

	if badge.RequiredBadge.Valid() {

		var user_valid bool = false

		user_badges := usr.Gaming.Badges

		for _, user_badge := range user_badges {

			if user_badge.Id == badge.RequiredBadge {

				user_valid = true
			}
		}

		if ! user_valid {

			return exceptions.OutOfBounds{"Don't have required badge."}
		}
	}

	badge_push := user.UserBadge{
		Id: id,
		Date: time.Now(),
	}

	err = database.C("users").Update(bson.M{"_id": usr.Id}, bson.M{"$push": bson.M{"gaming.badges": badge_push}})

	if err != nil {
		panic(err)
	}

	if badge.Coins > 0 {

		// Pay the badge price using user coins
		go self.Coins(-badge.Coins)
	}

	return nil
}