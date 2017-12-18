package user

import (
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"gopkg.in/mgo.v2/bson"

	"time"
)

type One struct {
	data *UserPrivate
	di   *Module
}

// Return data model
func (self *One) Data() *UserPrivate {
	return self.data
}

// Data update only persistent on runtime
func (self *One) RUpdate(data *UserPrivate) {
	self.data = data
}

func (self *One) Email() string {

	if self.data.Facebook != nil {

		fb := self.data.Facebook.(bson.M)

		if email, exists := fb["email"]; exists {

			return email.(string)
		}
	}

	return self.data.Email
}

func (self *One) Name() string {
	return self.data.UserName
}

// Helper method to track a signin from the user
func (self *One) TrackUserSignin(client_address string) {
	record := &CheckinModel{
		UserId:  self.data.Id,
		Address: client_address,
		Date:    time.Now(),
	}

	err := deps.Container.Mgo().C("checkins").Insert(record)

	if err != nil {
		panic(err)
	}
}

// Helper method to track a signin from the user
func (self *One) ROwns(entity string, id bson.ObjectId) {
	_, err := deps.Container.Mgo().C("user_owns").UpdateAll(
		bson.M{
			"related":    entity,
			"related_id": id,
			"user_id":    self.data.Id,
			"removed":    bson.M{"$exists": false},
		},
		bson.M{
			"$set": bson.M{"removed": true, "removed_at": time.Now()},
		},
	)

	if err != nil {
		panic(err)
	}
}

func (self *One) TrackView(entity string, entity_id bson.ObjectId) {
	database := deps.Container.Mgo()
	record := &ViewModel{
		UserId:    self.data.Id,
		Related:   entity,
		RelatedId: entity_id,
		Created:   time.Now(),
	}

	err := database.C("user_views").Insert(record)

	if err != nil {
		panic(err)
	}

	if entity == "component" {
		err := database.C("components").Update(bson.M{"_id": entity_id}, bson.M{"$inc": bson.M{"views": 1}})

		if err != nil {
			panic(err)
		}
	}
}

func (self *One) MarkAsValidated() {
	err := deps.Container.Mgo().C("users").Update(bson.M{"_id": self.data.Id}, bson.M{"$set": bson.M{"validated": true}})

	if err != nil {
		panic(err)
	}

	self.data.Validated = true

	// Confirm the referral in case it exists
	self.followReferral()
}

func (o *One) IsValidated() bool {
	return o.data.Validated
}

func (self *One) Update(data map[string]interface{}) (err error) {
	if password, exists := data["password"]; exists {
		data["password"] = helpers.Sha256(password.(string))
	}

	err = deps.Container.Mgo().C("users").Update(bson.M{"_id": self.data.Id}, bson.M{"$set": data})
	return
}

func (self *One) followReferral() {
	// Just update blindly
	deps.Container.Mgo().C("referrals").Update(bson.M{"user_id": self.data.Id}, bson.M{"$set": bson.M{"confirmed": true}})
}
