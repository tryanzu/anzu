package main

import (
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/core/mail"
	"github.com/tryanzu/core/deps"
)

func init() {

	// Run config service bootstraping sequences.
	config.Bootstrap()

	// Run dependencies bootstraping sequences.
	deps.Bootstrap()

	// Boot internal services.
	mail.Boot()
}
