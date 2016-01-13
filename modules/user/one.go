package user

import (
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type One struct {
	data *User
	di   *Module
}

// Return data model
func (self *One) Data() *User {
	return self.data
}

// Data update only persistent on runtime
func (self *One) RUpdate(data *User) {
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
	}

	return self
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

func (self *One) TrackView(entity string, entity_id bson.ObjectId) {

	database := self.di.Mongo.Database
	record := &ViewModel{
		UserId: self.data.Id,
		Related: entity,
		RelatedId: entity_id,
		Created: time.Now(),
	}

	err := database.C("user_views").Insert(record)

	if err != nil {
		panic(err)
	}
}

func (self *One) SendConfirmationEmail() {

	mailing := self.di.Mail

	compose := mail.Mail{
		Template: 250222,
		Recipient: []mail.MailRecipient{
			{
				Name:  self.data.UserName,
				Email: self.data.Email,
			},
		},
		Variables: map[string]interface{}{
			"confirm_url": "http://spartangeek.com/signup/confirm/" + self.data.VerificationCode,
		},
	}

	mailing.Send(compose)
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

func (self *One) followReferral() {

	di := self.di
	database := di.Mongo.Database

	// Just update blindly
	database.C("referrals").Update(bson.M{"user_id": self.data.Id}, bson.M{"$set": bson.M{"confirmed": true}})
}
