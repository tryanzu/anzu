package user

import (
	"errors"
	"html"
	"time"

	"github.com/tryanzu/core/core/config"
	"gopkg.in/mgo.v2/bson"
)

// ErrInvalidBanReason not present in config
var ErrInvalidBanReason = errors.New("invalid ban reason")

// ErrInvalidUser user cannot be found
var ErrInvalidUser = errors.New("invalid user to ban")

func ResetNotifications(d deps, id bson.ObjectId) (err error) {
	err = d.Mgo().C("users").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"notifications": 0}})
	return
}

// UpsertBan performs validations before upserting data struct
func UpsertBan(d deps, ban Ban) (Ban, error) {
	if ban.ID.Valid() == false {
		ban.ID = bson.NewObjectId()
		ban.Created = time.Now()
		ban.Status = ACTIVE
	}
	usr, err := FindId(d, ban.UserID)
	if err != nil {
		return ban, ErrInvalidUser
	}
	rules := config.C.Rules()
	rule, exists := rules.BanReasons[ban.Reason]
	if false == exists {
		return ban, ErrInvalidBanReason
	}
	effects, err := rule.Effects(ban.RelatedTo, usr.BannedTimes)
	if err != nil {
		panic(err)
	}
	mins := time.Minute * time.Duration(effects.Duration)
	ban.Until = time.Now().Add(mins)
	ban.Content = html.EscapeString(ban.Content)
	ban.Updated = time.Now()
	changes, err := d.Mgo().C("bans").UpsertId(ban.ID, bson.M{"$set": ban})
	if err != nil {
		return ban, err
	}
	if changes.Matched == 0 && ban.Status == ACTIVE {
		err = d.Mgo().C("users").UpdateId(ban.UserID, bson.M{
			"$set": bson.M{
				"banned_at":    ban.Created,
				"banned":       true,
				"banned_re":    ban.Reason,
				"banned_until": ban.Until,
			},
			"$inc": bson.M{
				"banned_times": 1,
			},
		})
		if err != nil {
			return ban, err
		}
		k := []byte("ban:")
		k = append(k, []byte(usr.Id)...)
		err = d.LedisDB().Set(k, []byte{})
		if err != nil {
			return ban, err
		}
		_, err = d.LedisDB().Expire(k, effects.Duration*60)
		if err != nil {
			return ban, err
		}
	}
	return ban, nil
}
