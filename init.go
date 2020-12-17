package main

import (
	"os"

	"github.com/facebookgo/inject"
	"github.com/getsentry/raven-go"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/facebook"
	"github.com/op/go-logging"
	"github.com/spf13/cobra"
	_ "github.com/tryanzu/core/board/events"
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/core/shell"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/acl"
	"github.com/tryanzu/core/modules/api"
	"github.com/tryanzu/core/modules/assets"
	"github.com/tryanzu/core/modules/exceptions"
	"github.com/tryanzu/core/modules/feed"
	"github.com/tryanzu/core/modules/gaming"
	"github.com/tryanzu/core/modules/notifications"
	"github.com/tryanzu/core/modules/security"
	"github.com/tryanzu/core/modules/user"
	"github.com/xuyu/goredis"
)

func main() {
	// Resources for the API
	var (
		api                 api.Module
		securityModule      security.Module
		notificationsModule notifications.NotificationsModule
		feedModule          feed.FeedModule
		exceptions          exceptions.ExceptionsModule
		log                 = logging.MustGetLogger("main")
	)

	// Services for the DI
	aclService := acl.Boot("roles.json")
	gamingService := gaming.Boot("gaming.json")
	userService := user.Boot()
	errorService, _ := raven.NewWithTags(deps.SentryURL, nil)
	cacheService, _ := goredis.Dial(&goredis.DialConfig{Address: deps.RedisURL})
	assetsService := assets.Boot()

	// Authentication services
	runtime := config.C.Copy()
	facebookCnf := runtime.Oauth.Facebook
	if len(facebookCnf.Key) > 0 && len(facebookCnf.Secret) > 0 {
		goth.UseProviders(facebook.New(facebookCnf.Key, facebookCnf.Secret, facebookCnf.Callback, "email"))
	}

	// Graph main object (used to inject dependencies)
	// LEGACY: To be deprecated asap
	var g inject.Graph
	err := g.Provide(
		&inject.Object{Value: log, Complete: true},
		&inject.Object{Value: errorService, Complete: true},
		&inject.Object{Value: cacheService, Complete: true},
		&inject.Object{Value: aclService, Complete: false},
		&inject.Object{Value: assetsService, Complete: false},
		&inject.Object{Value: userService, Complete: false},
		&inject.Object{Value: gamingService, Complete: false},
		&inject.Object{Value: &securityModule},
		&inject.Object{Value: &notificationsModule},
		&inject.Object{Value: &feedModule},
		&inject.Object{Value: &exceptions},
	)

	if err != nil {
		log.Fatal(err)
	}

	shellCmd := &cobra.Command{
		Use:   "shell",
		Short: "Starts interactive shell",
		Long: `Starts blacker interactive shell
		with helper tasks.
        `,
		Run: func(cmd *cobra.Command, args []string) {
			shell.RunShell()
		},
	}

	cmdAPI := &cobra.Command{
		Use:   "api",
		Short: "Starts API web server",
		Long: `Starts API web server listening
        in the specified env port
        `,
		Run: func(cmd *cobra.Command, args []string) {
			port := ":3200"
			if len(args) == 1 {
				port = args[0]
			}
			if v, exists := os.LookupEnv("BIND_TO"); exists {
				port = v
			}
			if v, exists := os.LookupEnv("PORT"); exists {
				port = v
			}

			// Populate dependencies using the already instantiated DI
			api.Populate(g)

			// Run API module
			api.Run(port)
		},
	}

	cmdSyncRanking := &cobra.Command{
		Use:   "sync-ranking",
		Short: "Sync ranking",
		Long: `Sync and recalculates ranking facts
		in proper manner
        `,
		Run: func(cmd *cobra.Command, args []string) {

			// Populate the DI with the instances
			if err := g.Populate(); err != nil {
				log.Fatal(err)
			}

			gamingService.ResetGeneralRanking()
		},
	}

	rootCmd := &cobra.Command{Use: "Anzu"}
	rootCmd.AddCommand(cmdAPI)
	rootCmd.AddCommand(cmdSyncRanking)
	rootCmd.AddCommand(shellCmd)
	rootCmd.Execute()
}
