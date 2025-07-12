package dal

import (
	"context"
	"time"

	"github.com/tryanzu/core/board/categories"
	"github.com/tryanzu/core/modules/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Seed(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parent := categories.Category{
		ID:          primitive.NewObjectID(),
		Name:        "General",
		Description: "All general matters can go here",
		Slug:        "general",
		Order:       0,
		Permissions: categories.ACL{
			Read:  []string{"*"},
			Write: []string{"*"},
		},
	}
	_, err := db.Collection("categories").InsertOne(ctx, parent)
	if err != nil {
		return err
	}
	category := categories.Category{
		ID:          primitive.NewObjectID(),
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
	_, err = db.Collection("categories").InsertOne(ctx, category)
	if err != nil {
		return err
	}
	_, err = user.InsertUser(db.Collection("users"), "admin", "admin", "admin@local.domain", user.Validated(true), user.WithRole("administrator"))
	if err != nil {
		return err
	}
	return nil
}
