package security

import (
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
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

		return !user_data.Banned
	}

	if ip.Banned == true && usr.Data().Banned == true {
		
		return false

	} else if ip.Banned == false && usr.Data().Banned == true {

		// In case the ip is not banned but the user is then update it
		err := database.C("trusted_addresses").Update(bson.M{"_id": ip.Id}, bson.M{"$set": bson.M{"banned": true}})

		if err != nil {
			panic(err)
		}

		return false

	} else if ip.Banned == true && usr.Data().Banned == false {

		// In case the ip is banned but the user is not then update it
		err := database.C("users").Update(bson.M{"_id": usr.Data().Id}, bson.M{"$set": bson.M{"banned": true}})

		if err != nil {
			panic(err)
		}

		return false
	}

	return true
}

func (module Module) TrustIP(address string) bool {

	var ip IpAddress

	database := module.Mongo.Database
	err := database.C("trusted_addresses").Find(bson.M{"address": address}).One(&ip)

	if err != nil {
		return true
	}

	if ip.Banned == true {
		return false
	}

	return true
}