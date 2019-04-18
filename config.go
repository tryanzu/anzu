package main

import (
	"log"

	"cloud.google.com/go/profiler"
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/core/mail"
	"github.com/tryanzu/core/core/templates"
	"github.com/tryanzu/core/deps"
)

func init() {

	// Run config service bootstraping sequences.
	config.Bootstrap()

	// Start profiling if enabled
	if c := config.C.Copy(); c.Profiler.Enabled == true {
		if err := profiler.Start(profiler.Config{
			Service:        "anzu",
			ServiceVersion: "0.1-alpha-rc1",
			ProjectID:      c.Profiler.Id, // optional on GCP
		}); err != nil {
			log.Fatalf("Cannot start the profiler: %v", err)
		}
	}

	// Run dependencies bootstraping sequences.
	deps.Bootstrap()

	// Boot internal services.
	mail.Boot()
	templates.Boot()
}
