package user

import (
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
		UserId: self.data.Id,
		Address: client_address,
		Date: time.Now(),
	}

	err := database.C("checkins").Insert(record)

	if err != nil {
		panic(err)
	}
}