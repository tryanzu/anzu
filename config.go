package main

import (
	"os"

	"github.com/subosito/gotenv"
	post "github.com/tryanzu/core/board/posts"
	"github.com/tryanzu/core/board/search"
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/core/content"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/mail"
	"github.com/tryanzu/core/core/templates"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/api"
)

func init() {
	gotenv.Load()
	if v, exists := os.LookupEnv("ENV"); exists {
		api.ENV = v
		deps.ENV = v
	}
	if v, exists := os.LookupEnv("NEW_RELIC_KEY"); exists {
		api.NewRelicKey = v
	}
	if v, exists := os.LookupEnv("NEW_RELIC_NAME"); exists {
		api.NewRelicName = v
	}
	if v, exists := os.LookupEnv("MONGO_URL"); exists {
		deps.MongoURL = v
	}
	if v, exists := os.LookupEnv("MONGO_NAME"); exists {
		deps.MongoName = v
	}
	if v, exists := os.LookupEnv("REDIS_URL"); exists {
		deps.RedisURL = v
	}
	if v, exists := os.LookupEnv("SENTRY_URL"); exists {
		deps.SentryURL = v
	}
	// Run config service bootstraping sequences.
	config.Bootstrap()
	events.Boot()

	// Run dependencies bootstraping sequences.
	deps.Bootstrap()

	// Boot internal services.
	content.Boot()
	post.Boot()
	mail.Boot()
	templates.Boot()
	search.Boot()
}
