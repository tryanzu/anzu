package acl

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"gopkg.in/mgo.v2/bson"
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

func (user *User) Can(permission string) bool {

	return user.isGranted(permission)
}

// Check if user can update post
func (user *User) CanUpdatePost(post model.Post) bool {

	return user.isActionGranted(post.UserId, post.Category, "edit-own-posts", "edit-board-posts", "edit-category-posts")
}

// Check if user can solve post
func (user *User) CanSolvePost(post model.Post) bool {

	return user.isActionGranted(post.UserId, post.Category, "solve-own-posts", "solve-board-posts", "solve-category-posts")
}

// Check if user can lock post
func (user *User) CanLockPost(post model.Post) bool {

	return user.isActionGranted(post.UserId, post.Category, "block-own-post-comments", "block-board-post-comments", "block-category-post-comments")
}

// Check if user can delete post
func (user *User) CanDeletePost(post model.Post) bool {

	return user.isActionGranted(post.UserId, post.Category, "edit-own-posts", "edit-board-posts", "edit-category-posts")
}

// Check if user can update comment
func (user *User) CanUpdateComment(comment model.Comment, post model.Post) bool {

	return user.isActionGranted(comment.UserId, post.Category, "edit-own-comments", "edit-board-comments", "edit-category-comments")
}

// Check if user can delete comment
func (user *User) CanDeleteComment(comment model.Comment, post model.Post) bool {

	return user.isActionGranted(comment.UserId, post.Category, "delete-own-comments", "delete-board-comments", "delete-category-comments")
}

// Check if user can do action, super_action or category_action
func (user *User) isActionGranted(entity_owner bson.ObjectId, category bson.ObjectId, action, super_action, category_action string) bool {

	// Post author check
	if entity_owner == user.data.Id {

		if user.isGranted(action) {

			return true
		}

	} else if entity_owner != user.data.Id {

		// Super ability to do action
		if user.isGranted(super_action) {

			return true
		}

		// Super ability to edit all the parent category posts
		if user.isGranted(category_action) {

			for _, role := range user.data.Roles {

				// Only allow updating when user has that category
				if allowed, _ := helpers.InArray(category, role.Categories); allowed {

					return true
				}
			}
		}
	}

	return false
}

// Check if permission is granted for all user roles
func (user *User) isGranted(permission string) bool {

	for _, role := range user.data.Roles {

		p, exists := user.acl.Permissions[permission]

		if exists {
			if user.acl.Map.IsGranted(role.Name, p, nil) {
				// User's role is granted to do "permission"
				return true
			}
		}
	}

	return false
}

func (user *User) checkRolesRecursive(name string, compare string) bool {

	_, parents, err := user.acl.Map.Get(name)

	if err == nil {

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
