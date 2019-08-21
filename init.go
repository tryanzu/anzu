package main

import (
	"fmt"
	"os"

	"github.com/facebookgo/inject"
	"github.com/getsentry/raven-go"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/facebook"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/op/go-logging"
	"github.com/spf13/cobra"
	_ "github.com/tryanzu/core/board/events"
	"github.com/tryanzu/core/core/shell"
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

	// Graph main object (used to inject dependencies)
	var g inject.Graph

	// Run with the specified env file
	envfile := os.Getenv("ENV_FILE")
	if envfile == "" {
		envfile = "./env.json"
	}

	// Resources for the API
	var (
		api                 api.Module
		securityModule      security.Module
		notificationsModule notifications.NotificationsModule
		feedModule          feed.FeedModule
		exceptions          exceptions.ExceptionsModule
		log                 = logging.MustGetLogger("blacker")
		format              = logging.MustStringFormatter(
			`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
		)
	)

	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)

	// Services for the DI
	configService, _ := config.ParseJsonFile(envfile)
	aclService := acl.Boot(string_value(configService.String("application.acl")))
	gamingService := gaming.Boot(string_value(configService.String("application.gaming")))
	userService := user.Boot()
	errorService, _ := raven.NewClient(string_value(configService.String("sentry.dns")), nil)
	cacheService, _ := goredis.Dial(&goredis.DialConfig{Address: string_value(configService.String("cache.redis"))})
	assetsService := assets.Boot()

	// Authentication services
	facebookProvider := facebook.New(string_value(configService.String("auth.facebook.key")), string_value(configService.String("auth.facebook.secret")), string_value(configService.String("auth.facebook.callback")), "email")
	fmt.Printf("facebook provider client %s secret %s", facebookProvider.ClientKey, facebookProvider.Secret)
	facebookProvider.Debug(true)
	goth.UseProviders(facebookProvider)

	// Amazon services for the DI
	amazonAuth, err := aws.GetAuth(string_value(configService.String("amazon.access_key")), string_value(configService.String("amazon.secret")))
	if err != nil {
		panic(err)
	}

	s3Region := aws.USWest
	s3Service := s3.New(amazonAuth, s3Region)
	s3BucketService := s3Service.Bucket(string_value(configService.String("amazon.s3.bucket")))

	// Provide graph with service instances
	err = g.Provide(
		&inject.Object{Value: log, Complete: true},
		&inject.Object{Value: configService, Complete: true},
		&inject.Object{Value: errorService, Complete: true},
		&inject.Object{Value: cacheService, Complete: true},
		&inject.Object{Value: s3Service, Complete: true},
		&inject.Object{Value: s3BucketService, Complete: true},
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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

	var cmdAPI = &cobra.Command{
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

			// Populate dependencies using the already instantiated DI
			api.Populate(g)

			// Run API module
			api.Run(port)
		},
	}

	var cmdSyncRanking = &cobra.Command{
		Use:   "sync-ranking",
		Short: "Sync ranking",
		Long: `Sync and recalculates ranking facts
		in proper manner
        `,
		Run: func(cmd *cobra.Command, args []string) {

			// Populate the DI with the instances
			if err := g.Populate(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			gamingService.ResetGeneralRanking()
		},
	}

	var rootCmd = &cobra.Command{Use: "Anzu"}
	rootCmd.AddCommand(cmdAPI)
	rootCmd.AddCommand(cmdSyncRanking)
	rootCmd.AddCommand(shellCmd)
	rootCmd.Execute()

	return
}

func string_value(value string, err error) string {
	if err != nil {
		panic(err)
	}

	return value
}
