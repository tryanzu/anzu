package deps

import (
	"context"
	"flag"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// MongoURL config uri
	MongoURL string
	// MongoName db name
	MongoName string

	ShouldSeed = flag.Bool("should-seed", false, "determines whether we seed the initial categories and admin user to bootstrap the site")
)

func IgniteMongoDB(container Deps) (Deps, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURL))
	if err != nil {
		log.Error(err)
		log.Info(MongoURL)
		return container, err
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Error(err)
		return container, err
	}

	db := client.Database(MongoName)
	collections, err := db.ListCollectionNames(ctx, map[string]any{})
	if err != nil {
		return container, err
	}
	seed := !slices.Contains(collections, "users")
	if seed {
		ShouldSeed = &seed
	}

	// Ensure indexes
	usersCol := db.Collection("users")

	// Email index
	emailIndexModel := mongo.IndexModel{
		Keys:    map[string]any{"email": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err = usersCol.Indexes().CreateOne(ctx, emailIndexModel)
	if err != nil {
		log.Error("Failed to create email index:", err)
	}

	// Username index
	usernameIndexModel := mongo.IndexModel{
		Keys:    map[string]any{"username": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err = usersCol.Indexes().CreateOne(ctx, usernameIndexModel)
	if err != nil {
		log.Error("Failed to create username index:", err)
	}

	// Text search index for posts
	postsCol := db.Collection("posts")
	searchIndexModel := mongo.IndexModel{
		Keys: map[string]any{
			"title":   "text",
			"content": "text",
		},
	}
	_, err = postsCol.Indexes().CreateOne(ctx, searchIndexModel)
	if err != nil {
		log.Error("Failed to create text search index:", err)
	}

	container.DatabaseSessionProvider = client
	container.DatabaseProvider = db

	return container, nil
}
