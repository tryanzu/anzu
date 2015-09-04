package user

type One struct {
	data *User
}

// Return data model
func (self *One) Data() *User {
	return self.data
}

// Data update only persistent on runtime
func (self *One) RUpdate(data *User) {
	self.data = data
}
