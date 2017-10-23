package main

import (
	"fmt"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/facebookgo/inject"
	_ "github.com/fernandez14/spartangeek-blacker/board/events"
	"github.com/fernandez14/spartangeek-blacker/core/shell"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/fernandez14/spartangeek-blacker/modules/api"
	"github.com/fernandez14/spartangeek-blacker/modules/assets"
	"github.com/fernandez14/spartangeek-blacker/modules/cli"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"github.com/fernandez14/spartangeek-blacker/modules/notifications"
	"github.com/fernandez14/spartangeek-blacker/modules/security"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/getsentry/raven-go"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/facebook"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/op/go-logging"
	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	"github.com/xuyu/goredis"

	"os"
	"runtime"
)

func main() {

	// Start by using the power of the machine cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Run dependencies bootstraping sequences.
	deps.Bootstrap()

	// Graph main object (used to inject dependencies)
	var g inject.Graph

	// Run with the specified env file
	envfile := os.Getenv("ENV_FILE")
	if envfile == "" {
		envfile = "./env.json"
	}

	// Resources for the API
	var api api.Module
	var cliModule cli.Module
	var securityModule security.Module
	var notificationsModule notifications.NotificationsModule
	var feedModule feed.FeedModule
	var exceptions exceptions.ExceptionsModule
	var log *logging.Logger = logging.MustGetLogger("blacker")
	var format logging.Formatter = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)

	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)

	// Services for the DI
	configService, _ := config.ParseJsonFile(envfile)
	aclService := acl.Boot(string_value(configService.String("application.acl")))
	gamingService := gaming.Boot(string_value(configService.String("application.gaming")))
	userService := user.Boot()
	componentsService := components.Boot()
	mongoService := mongo.NewService(string_value(configService.String("database.uri")), string_value(configService.String("database.name")))
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

	// Statsd - Tracking
	prefix := "blacker."
	statsService, err := statsd.NewClient("127.0.0.1:8125", prefix)
	if err != nil {
		panic(err)
	}

	// Provide graph with service instances
	err = g.Provide(
		&inject.Object{Value: log, Complete: true},
		&inject.Object{Value: configService, Complete: true},
		&inject.Object{Value: mongoService, Complete: true},
		&inject.Object{Value: errorService, Complete: true},
		&inject.Object{Value: cacheService, Complete: true},
		&inject.Object{Value: s3Service, Complete: true},
		&inject.Object{Value: s3BucketService, Complete: true},
		&inject.Object{Value: statsService, Complete: true},
		&inject.Object{Value: aclService, Complete: false},
		&inject.Object{Value: assetsService, Complete: false},
		&inject.Object{Value: userService, Complete: false},
		&inject.Object{Value: componentsService, Complete: false},
		&inject.Object{Value: gamingService, Complete: false},
		&inject.Object{Value: &cliModule},
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

	var cmdApi = &cobra.Command{
		Use:   "api",
		Short: "Starts API web server",
		Long: `Starts API web server listening
        in the specified env port
        `,
		Run: func(cmd *cobra.Command, args []string) {

			// Populate dependencies using the already instantiated DI
			api.Populate(g)

			// Run API module
			api.Run()
		},
	}

	var cmdJobs = &cobra.Command{
		Use:   "jobs",
		Short: "Starts Jobs worker",
		Long: `Starts jobs worker daemon
        so things can run like crons
        `,
		Run: func(cmd *cobra.Command, args []string) {

			// Populate the DI with the instances
			if err := g.Populate(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			// Start the jobs
			c := cron.New()

			// Reset the user temporal stuff each X
			c.AddFunc("@midnight", gamingService.ResetTempStuff)
			c.AddFunc("@every 8h", gamingService.ResetGeneralRanking)

			// Start the jobs
			c.Start()

			select {}
		},
	}

	var cmdSyncGamification = &cobra.Command{
		Use:   "sync-gamification",
		Short: "Sync gamification",
		Long: `Sync and recalculates gamification facts
		in proper manner
        `,
		Run: func(cmd *cobra.Command, args []string) {

			// Populate the DI with the instances
			if err := g.Populate(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			gamingService.ResetTempStuff()
		},
	}

	var cmdRunRoutine = &cobra.Command{
		Use:   "run [routine]",
		Short: "Run cli routine",
		Long: `Run specified routine
		from cli module`,
		Run: func(cmd *cobra.Command, args []string) {

			// Populate the DI with the instances
			if err := g.Populate(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			cliModule.Run(args[0])
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

	var rootCmd = &cobra.Command{Use: "blacker"}
	rootCmd.AddCommand(cmdApi)
	rootCmd.AddCommand(cmdSyncGamification)
	rootCmd.AddCommand(cmdSyncRanking)
	rootCmd.AddCommand(cmdJobs)
	rootCmd.AddCommand(cmdRunRoutine)
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
