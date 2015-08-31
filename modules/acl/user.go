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

	// Iterate over each of the user roles
	for _, role := range roles {

		// Basic check of the existence of the role inside the category
		if allowed, _ := helpers.InArray(role.Name, category.Permissions.Write); allowed {
			return true
		}

		// Deep check within parents
		for _, allowed := range category.Permissions.Write {

			if user.checkRolesRecursive(role.Name, allowed) {
				return true
			}
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

		if allowed, _ := helpers.InArray(role.Name, category.Permissions.Read); allowed {
			return true
		}
	}

	return false
}

func (user *User) checkRolesRecursive(role string, compare string) bool {

	if role_permissions := user.acl.Map.Get(role); role_permissions != nil {

		parents := role_permissions.Parents()
		
		if allowed, _ := helpers.InArray(compare, parents); allowed {
			return true
		}

		for _, parent_role := range parents {

			allowed := user.checkRolesRecursive(parent_role, compare)

			if allowed {
				return true
			}
		}
	}

	return false
}
