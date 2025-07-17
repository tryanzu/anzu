package security

import (
	"context"
	"time"

	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/user"
	"github.com/xuyu/goredis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Module struct {
	Redis *goredis.Redis `inject:""`
}

func (module Module) TrustUserIP(address string, usr *user.One) bool {
	ctx := context.Background()
	var (
		ip  IpAddress
		err error
	)
	database := deps.Container.Mgo()
	user := usr.Data()

	// The address haven't been trusted before so we need to lookup
	trustedAddressesCollection := database.Collection("trusted_addresses")
	err = trustedAddressesCollection.FindOne(ctx, bson.M{"address": address}).Decode(&ip)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			trusted := &IpAddress{
				Address: address,
				Users:   []primitive.ObjectID{user.Id},
				Banned:  user.Banned,
			}
			_, err = trustedAddressesCollection.InsertOne(ctx, trusted)
			return err != nil && !user.Banned
		}
		return false
	}

	if ip.Banned && user.Banned {
		return false
	} else if !ip.Banned && user.Banned {
		// In case the ip is not banned but the user is then update it
		filter := bson.M{"_id": ip.Id}
		update := bson.M{"$set": bson.M{
			"banned":    true,
			"banned_at": time.Now(),
		}, "$push": bson.M{"banned_reason": user.UserName + " has propagated the ban to the IP address."}}
		_, err = trustedAddressesCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			panic(err)
		}
		return false
	} else if ip.Banned && !user.Banned {
		// In case the ip is banned but the user is not then update it
		usersCollection := database.Collection("users")
		filter := bson.M{"_id": user.Id}
		update := bson.M{"$set": bson.M{"banned": true, "banned_at": time.Now()}, "$push": bson.M{"banned_reason": user.UserName + " has accessed from a flagged IP. " + ip.Address}}
		_, err = usersCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			panic(err)
		}
		return false
	}

	return true
}

func (module Module) TrustIP(address string) bool {
	ctx := context.Background()
	var ip IpAddress
	database := deps.Container.Mgo()
	collection := database.Collection("trusted_addresses")
	err := collection.FindOne(ctx, bson.M{"address": address}).Decode(&ip)

	if err != nil {
		return err == mongo.ErrNoDocuments
	}

	if ip.Banned {
		return false
	}

	return true
}
