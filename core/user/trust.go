package user

func CanBeTrusted(user User) bool {
	return user.Warnings < 6
}
