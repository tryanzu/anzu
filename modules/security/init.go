package security

import (
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/xuyu/goredis"
)

type Module struct {
	Mongo *mongo.Service `inject:""`
	Redis *goredis.Redis `inject:""`
}

func (module Module) TrustUserIP(address string, usr *user.One) bool {

	var ip IpAddress

	database := module.Mongo.Database
	err := database.C("trusted_addresses").Find(bson.M{"address": address}).One(&ip)

	if err != nil {
		
		user_data := usr.Data()

		// The address haven't been trusted before so we need to lookup 
		trusted := &IpAddress{
			Address: address,
			Users: []bson.ObjectId{user_data.Id},
			Banned: user_data.Banned,
		}

		err := database.C("trusted_addresses").Insert(trusted)

		if err != nil {
			return false
		}

		return user_data.Banned
	}

	if ip.Banned == true {
		return false
	}

	return true
}