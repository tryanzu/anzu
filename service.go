package main

import (
	"fmt"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/cosn/firebase"
	"github.com/facebookgo/inject"
	"github.com/fernandez14/spartangeek-blacker/interfaces"
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/fernandez14/spartangeek-blacker/modules/api"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"github.com/fernandez14/spartangeek-blacker/modules/notifications"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/getsentry/raven-go"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	"github.com/xuyu/goredis"
	"os"
	"runtime"
)

func main() {

	// Graph main object (used to inject dependencies)
	var g inject.Graph

	// Start by using the power of the machine cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Run with the specified env file
	envfile := os.Getenv("ENV_FILE")
	if envfile == "" {
		envfile = "./env.json"
	}

	// Resources for the API
	var api api.Module
	var notificationsModule notifications.NotificationsModule
	var feedModule feed.FeedModule
	var exceptions exceptions.ExceptionsModule

	// Services for the DI
	configService, _ := config.ParseJsonFile(envfile)
	aclService := acl.Boot(string_value(configService.String("application.acl")))
	gamingService := gaming.Boot(string_value(configService.String("application.gaming")))
	userService := user.Boot()
	mongoService := mongo.NewService(string_value(configService.String("database.uri")), string_value(configService.String("database.name")))
	errorService, _ := raven.NewClient(string_value(configService.String("sentry.dns")), nil)
	cacheService, _ := goredis.Dial(&goredis.DialConfig{Address: string_value(configService.String("cache.redis"))})
	firebaseService := new(firebase.Client)
	firebaseService.Init(string_value(configService.String("firebase.url")), string_value(configService.String("firebase.secret")), nil)

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

	// Implementations will be fullfilled manually
	firebaseBroadcaster := notifications.FirebaseBroadcaster{Firebase: firebaseService}
	broadcaster := interfaces.NotificationBroadcaster(firebaseBroadcaster)

	// Provide graph with service instances
	err = g.Provide(
		&inject.Object{Value: configService, Complete: true},
		&inject.Object{Value: mongoService, Complete: true},
		&inject.Object{Value: errorService, Complete: true},
		&inject.Object{Value: cacheService, Complete: true},
		&inject.Object{Value: s3Service, Complete: true},
		&inject.Object{Value: s3BucketService, Complete: true},
		&inject.Object{Value: firebaseService, Complete: true},
		&inject.Object{Value: statsService, Complete: true},
		&inject.Object{Value: aclService, Complete: false},
		&inject.Object{Value: userService, Complete: false},
		&inject.Object{Value: gamingService, Complete: false},
		&inject.Object{Value: broadcaster, Complete: true, Name: "Notifications"},
		&inject.Object{Value: &notificationsModule},
		&inject.Object{Value: &feedModule},
		&inject.Object{Value: &exceptions},
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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

			// Start the jobs
			c := cron.New()

			// Reset the user temporal stuff each X
			c.AddFunc("@midnight", gamingService.ResetTempStuff)

			// Start the jobs
			c.Start()

			select {}
		},
	}

	var rootCmd = &cobra.Command{Use: "blacker"}
	rootCmd.AddCommand(cmdApi)
	rootCmd.AddCommand(cmdJobs)
	rootCmd.Execute()

	return
}

func string_value(value string, err error) string {

	if err != nil {
		panic(err)
	}

	return value
}
