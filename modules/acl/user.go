package acl

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
)

type User struct {
	data model.User  
	acl  *Module
}

func (user *User) CanWrite(category model.Category) bool {


	if allowed, _ := helpers.InArray("*", category.Permissions.Write); allowed {
		return true
	}

	roles := user.data.Roles

	for _, role := range roles {

		if allowed, _ := helpers.InArray(role, category.Permissions.Write); allowed {
			return true
		}	
	}

	return false
}

func (user *User) CanRead(category model.Category) bool {

	if allowed, _ := helpers.InArray("*", category.Permissions.Read); allowed {
		return true
	}

	roles := user.data.Roles

	for _, role := range roles {

		if allowed, _ := helpers.InArray(role, category.Permissions.Read); allowed {
			return true
		}	
	}

	return false
}
