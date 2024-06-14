package dal

import (
	"github.com/tryanzu/core/board/categories"
	"github.com/tryanzu/core/modules/user"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func Seed(db *mgo.Database) error {
	parent := categories.Category{
		ID:          bson.NewObjectId(),
		Name:        "General",
		Description: "All general matters can go here",
		Slug:        "general",
		Order:       0,
		Permissions: categories.ACL{
			Read:  []string{"*"},
			Write: []string{"*"},
		},
	}
	err := db.C("categories").Insert(parent)
	if err != nil {
		return err
	}
	category := categories.Category{
		ID:          bson.NewObjectId(),
		Parent:      parent.ID,
		Name:        "General",
		Description: "All general matters can go here",
		Slug:        "general",
		Order:       0,
		Permissions: categories.ACL{
			Read:  []string{"*"},
			Write: []string{"*"},
		},
	}
	err = db.C("categories").Insert(category)
	if err != nil {
		return err
	}
	_, err = user.InsertUser(db.C("users"), "admin", "admin", "admin@local.domain", user.Validated(true), user.WithRole("administrator"))
	if err != nil {
		return err
	}
	return nil
}
