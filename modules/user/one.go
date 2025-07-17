package user

import (
	"context"
	"time"

	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	ctx := context.Background()
	record := &CheckinModel{
		UserId:  self.data.Id,
		Address: client_address,
		Date:    time.Now(),
	}

	database := deps.Container.Mgo()
	collection := database.Collection("checkins")
	_, err := collection.InsertOne(ctx, record)
	if err != nil {
		panic(err)
	}
}

// Helper method to track a signin from the user
func (self *One) ROwns(entity string, id primitive.ObjectID) {
	ctx := context.Background()
	database := deps.Container.Mgo()
	collection := database.Collection("user_owns")
	filter := bson.M{
		"related":    entity,
		"related_id": id,
		"user_id":    self.data.Id,
		"removed":    bson.M{"$exists": false},
	}
	update := bson.M{
		"$set": bson.M{"removed": true, "removed_at": time.Now()},
	}
	_, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		panic(err)
	}
}

func (self *One) TrackView(entity string, entity_id primitive.ObjectID) {
	ctx := context.Background()
	database := deps.Container.Mgo()
	record := &ViewModel{
		UserId:    self.data.Id,
		Related:   entity,
		RelatedId: entity_id,
		Created:   time.Now(),
	}

	userViewsCollection := database.Collection("user_views")
	_, err := userViewsCollection.InsertOne(ctx, record)
	if err != nil {
		panic(err)
	}

	if entity == "component" {
		componentsCollection := database.Collection("components")
		filter := bson.M{"_id": entity_id}
		update := bson.M{"$inc": bson.M{"views": 1}}
		_, err := componentsCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			panic(err)
		}
	}
}

func (self *One) MarkAsValidated() {
	ctx := context.Background()
	database := deps.Container.Mgo()
	collection := database.Collection("users")
	filter := bson.M{"_id": self.data.Id}
	update := bson.M{"$set": bson.M{"validated": true}}
	_, err := collection.UpdateOne(ctx, filter, update)
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
	ctx := context.Background()
	if password, exists := data["password"]; exists {
		data["password"] = helpers.Sha256(password.(string))
	}

	database := deps.Container.Mgo()
	collection := database.Collection("users")
	filter := bson.M{"_id": self.data.Id}
	update := bson.M{"$set": data}
	_, err = collection.UpdateOne(ctx, filter, update)
	return
}

func (self *One) followReferral() {
	ctx := context.Background()
	// Just update blindly
	database := deps.Container.Mgo()
	collection := database.Collection("referrals")
	filter := bson.M{"user_id": self.data.Id}
	update := bson.M{"$set": bson.M{"confirmed": true}}
	_, _ = collection.UpdateOne(ctx, filter, update)
}
