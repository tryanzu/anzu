package user

import (
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"gopkg.in/mgo.v2/bson"

	"errors"
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

// Load data helper
func (self *One) Load(section string) *One {

	switch section {
	case "referrals":

		var users_id []bson.ObjectId
		var referrals []ReferralModel
		var users []UserLightModel

		di := self.di
		database := di.Mongo.Database

		count, err := database.C("referrals").Find(bson.M{"owner_id": self.data.Id, "confirmed": true}).Count()

		if err != nil {
			panic(err)
		}

		err = database.C("referrals").Find(bson.M{"owner_id": self.data.Id, "confirmed": true}).Limit(10).Sort("-created_at").All(&referrals)

		if err != nil {
			panic(err)
		}

		for _, referral := range referrals {

			users_id = append(users_id, referral.UserId)
		}

		err = database.C("users").Find(bson.M{"_id": bson.M{"$in": users_id}}).Select(bson.M{"_id": 1, "username": 1, "email": 1, "image": 1}).All(&users)

		self.data.Referrals = ReferralsModel{
			Count: count,
			List:  users,
		}

	case "components":

		self.loadOwnedComponents()
	}

	return self
}

func (self *One) loadOwnedComponents() {

	var ownings []OwnModel
	var components_list []bson.ObjectId
	var components []OwnedComponent

	di := self.di
	database := di.Mongo.Database
	err := database.C("user_owns").Find(bson.M{"user_id": self.data.Id, "related": "component", "removed": bson.M{"$exists": false}}).Sort("created_at").All(&ownings)

	if err != nil {
		panic(err)
	}

	for _, owning := range ownings {
		components_list = append(components_list, owning.RelatedId)
	}

	err = database.C("components").Find(bson.M{"_id": bson.M{"$in": components_list}}).All(&components)

	if err != nil {
		panic(err)
	}

	for index, component := range components {

		for _, owning := range ownings {

			if component.Id == owning.RelatedId {

				components[index].Relationship = OwnRelationship{
					Type:    owning.Type,
					Created: owning.Created,
				}

				break
			}
		}
	}

	self.data.Components = components
}

// Helper method to track a signin from the user
func (self *One) TrackUserSignin(client_address string) {

	di := self.di
	database := di.Mongo.Database

	record := &CheckinModel{
		UserId:  self.data.Id,
		Address: client_address,
		Date:    time.Now(),
	}

	err := database.C("checkins").Insert(record)

	if err != nil {
		panic(err)
	}
}

// Helper method to track a signin from the user
func (self *One) Owns(status *string, entity string, id bson.ObjectId) {

	self.ROwns(entity, id)

	if status != nil {
		di := self.di
		database := di.Mongo.Database

		record := &OwnModel{
			UserId:    self.data.Id,
			Related:   entity,
			RelatedId: id,
			Type:      *status,
			Created:   time.Now(),
		}

		err := database.C("user_owns").Insert(record)

		if err != nil {
			panic(err)
		}
	}
}

// Helper method to track a signin from the user
func (self *One) ROwns(entity string, id bson.ObjectId) {

	di := self.di
	database := di.Mongo.Database

	_, err := database.C("user_owns").UpdateAll(bson.M{"related": entity, "related_id": id, "user_id": self.data.Id, "removed": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"removed": true, "removed_at": time.Now()}})

	if err != nil {
		panic(err)
	}
}

func (self *One) GetVoteStatus(name string, id bson.ObjectId) (string, error) {

	var model OwnModel

	di := self.di
	database := di.Mongo.Database
	err := database.C("user_owns").Find(bson.M{"related": name, "related_id": id, "user_id": self.data.Id, "removed": bson.M{"$exists": false}}).Sort("-created_at").One(&model)

	if err != nil {
		return "", errors.New("not-available")
	}

	return model.Type, nil
}

func (self *One) TrackView(entity string, entity_id bson.ObjectId) {

	database := self.di.Mongo.Database
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

	di := self.di
	database := di.Mongo.Database

	err := database.C("users").Update(bson.M{"_id": self.data.Id}, bson.M{"$set": bson.M{"validated": true}})

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

func (self *One) Update(data map[string]interface{}) error {

	database := self.di.Mongo.Database

	if password, exists := data["password"]; exists {
		data["password"] = helpers.Sha256(password.(string))
	}

	err := database.C("users").Update(bson.M{"_id": self.data.Id}, bson.M{"$set": data})

	return err
}

func (self *One) followReferral() {

	di := self.di
	database := di.Mongo.Database

	// Just update blindly
	database.C("referrals").Update(bson.M{"user_id": self.data.Id}, bson.M{"$set": bson.M{"confirmed": true}})
}
