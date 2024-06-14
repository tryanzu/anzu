package security

import (
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/user"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"

	"time"
)

type Module struct {
	Redis *goredis.Redis `inject:""`
}

func (module Module) TrustUserIP(address string, usr *user.One) bool {
	var (
		ip  IpAddress
		err error
	)
	mgo := deps.Container.Mgo()
	user := usr.Data()

	// The address haven't been trusted before so we need to lookup
	err = mgo.C("trusted_addresses").Find(bson.M{"address": address}).One(&ip)
	if err != nil {
		trusted := &IpAddress{
			Address: address,
			Users:   []bson.ObjectId{user.Id},
			Banned:  user.Banned,
		}
		err = mgo.C("trusted_addresses").Insert(trusted)
		return err != nil && !user.Banned
	}

	if ip.Banned && user.Banned {
		return false
	} else if !ip.Banned && user.Banned {

		// In case the ip is not banned but the user is then update it
		err = mgo.C("trusted_addresses").Update(
			bson.M{"_id": ip.Id},
			bson.M{"$set": bson.M{
				"banned":    true,
				"banned_at": time.Now(),
			}, "$push": bson.M{"banned_reason": user.UserName + " has propagated the ban to the IP address."}},
		)
		if err != nil {
			panic(err)
		}
		return false
	} else if ip.Banned && !user.Banned {

		// In case the ip is banned but the user is not then update it
		err = mgo.C("users").Update(bson.M{"_id": user.Id}, bson.M{"$set": bson.M{"banned": true, "banned_at": time.Now()}, "$push": bson.M{"banned_reason": user.UserName + " has accessed from a flagged IP. " + ip.Address}})
		if err != nil {
			panic(err)
		}

		return false
	}

	return true
}

func (module Module) TrustIP(address string) bool {
	var ip IpAddress
	err := deps.Container.Mgo().C("trusted_addresses").Find(bson.M{"address": address}).One(&ip)

	if err != nil {
		return true
	}

	if ip.Banned {
		return false
	}

	return true
}
