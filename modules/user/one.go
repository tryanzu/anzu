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

func (self *One) SendConfirmationEmail() {

	mailing := self.di.Mail

	compose := mail.Mail{
		Subject: "Último paso para tu registro",
		Template: "signup",
		Recipient: []mail.MailRecipient{
			{
				Name: self.data.UserName,
				Email: self.data.Email,
			},
		},
		Variables: map[string]string{
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