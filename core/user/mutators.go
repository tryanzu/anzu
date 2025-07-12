package user

import (
	"context"
	"errors"
	"html"
	"time"

	"github.com/tryanzu/core/core/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrInvalidBanReason not present in config
var ErrInvalidBanReason = errors.New("invalid ban reason")

// ErrInvalidUser user cannot be found
var ErrInvalidUser = errors.New("invalid user to ban")

func ResetNotifications(d deps, id primitive.ObjectID) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = d.Mgo().Collection("users").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"notifications": 0}})
	return
}

// LastSeenAt mutation
func LastSeenAt(d deps, id primitive.ObjectID, t time.Time) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = d.Mgo().Collection("users").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"last_seen_at": t}})
	return
}

// UpsertBan performs validations before upserting data struct
func UpsertBan(d deps, ban Ban) (Ban, error) {
	if ban.ID.IsZero() {
		ban.ID = primitive.NewObjectID()
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := options.Update().SetUpsert(true)
	result, err := d.Mgo().Collection("bans").UpdateOne(ctx, bson.M{"_id": ban.ID}, bson.M{"$set": ban}, opts)
	if err != nil {
		return ban, err
	}
	if result.MatchedCount == 0 && ban.Status == ACTIVE {
		_, err = d.Mgo().Collection("users").UpdateOne(ctx, bson.M{"_id": ban.UserID}, bson.M{
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
		k = append(k, []byte(usr.Id.Hex())...)
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

// UseRecoveryToken to generate auth token.
func UseRecoveryToken(d deps, clientIP, token string) (user User, jwtAuthToken string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = d.Mgo().Collection("user_recovery_tokens").UpdateOne(ctx,
		bson.M{
			"token":      token,
			"used":       false,
			"created_at": bson.M{"$gte": time.Now().Add(-15 * time.Minute)},
		},
		bson.M{
			"$set": bson.M{
				"used_at": time.Now(),
				"used":    true,
			},
		},
	)
	if err != nil {
		return
	}
	var t recoveryToken
	err = d.Mgo().Collection("user_recovery_tokens").FindOne(ctx, bson.M{"token": token}).Decode(&t)
	if err != nil {
		return
	}
	user, err = FindId(d, t.UserID)
	if err != nil {
		return
	}
	jwtAuthToken = genToken(clientIP, t.UserID, user.Roles, 1)
	return
}
