package main

import (
	"fmt"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/cosn/firebase"
	"github.com/facebookgo/inject"
	"github.com/fernandez14/go-siftscience"
	"github.com/fernandez14/spartangeek-blacker/interfaces"
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/fernandez14/spartangeek-blacker/modules/api"
	"github.com/fernandez14/spartangeek-blacker/modules/assets"
	"github.com/fernandez14/spartangeek-blacker/modules/cli"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/notifications"
	"github.com/fernandez14/spartangeek-blacker/modules/payments"
	"github.com/fernandez14/spartangeek-blacker/modules/preprocessor"
	"github.com/fernandez14/spartangeek-blacker/modules/queue"
	"github.com/fernandez14/spartangeek-blacker/modules/search"
	"github.com/fernandez14/spartangeek-blacker/modules/security"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/fernandez14/spartangeek-blacker/modules/transmit"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/getsentry/raven-go"
	"github.com/leebenson/paypal"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go/client"
	"github.com/xuyu/goredis"
	"gopkg.in/op/go-logging.v1"
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
	var transmitModule transmit.Module
	var preprocessor preprocessor.Module
	var cliModule cli.Module
	var queueModule queue.Module
	var securityModule security.Module
	var notificationsModule notifications.NotificationsModule
	var feedModule feed.FeedModule
	var exceptions exceptions.ExceptionsModule
	var gcommerceModule gcommerce.Module

	var log *logging.Logger = logging.MustGetLogger("blacker")
	var format logging.Formatter = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)

	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)

	storeService := store.Boot()

	// Services for the DI
	configService, _ := config.ParseJsonFile(envfile)
	aclService := acl.Boot(string_value(configService.String("application.acl")))
	gamingService := gaming.Boot(string_value(configService.String("application.gaming")))
	userService := user.Boot()
	componentsService := components.Boot()
	mongoService := mongo.NewService(string_value(configService.String("database.uri")), string_value(configService.String("database.name")))
	errorService, _ := raven.NewClient(string_value(configService.String("sentry.dns")), nil)
	cacheService, _ := goredis.Dial(&goredis.DialConfig{Address: string_value(configService.String("cache.redis"))})
	firebaseService := new(firebase.Client)
	firebaseService.Init(string_value(configService.String("firebase.url")), string_value(configService.String("firebase.secret")), nil)

	mailConfig, err := configService.Get("mail")

	if err != nil {
		panic(err)
	}

	gosift.ApiKey = string_value(configService.String("ecommerce.siftscience.api_key"))
	searchConfig, err := configService.Get("algolia")

	if err != nil {
		panic(err)
	}

	searchService := search.Boot(searchConfig)
	assetsService := assets.Boot()
	transmitService := transmit.Boot(string_value(configService.String("zmq.push")))
	mailService := mail.Boot(string_value(configService.String("mail.api_key")), mailConfig, false)

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

	// Payments module
	paymentGateways := map[string]payments.Gateway{}
	paypalConfig, err := configService.Get("ecommerce.paypal")
	{
		if err != nil {
			panic("Could not get paypal configuration to initialize payments module.")
		}

		clientID, err := paypalConfig.String("clientID")

		if err != nil {
			panic("Could not get config data to initialize paypal client. (cid)")
		}

		secret, err := paypalConfig.String("secret")

		if err != nil {
			panic("Could not get config data to initialize paypal client. (secret)")
		}

		sandbox, err := paypalConfig.Bool("sandbox")

		if err != nil {
			panic("Could not get config data to initialize paypal client. (sd)")
		}

		var r string = paypal.APIBaseSandBox

		if !sandbox {
			r = paypal.APIBaseLive
		}

		paypalClient := paypal.NewClient(clientID, secret, r)
		paypalGateway := &payments.Paypal{
			Client: paypalClient,
		}

		paymentGateways["paypal"] = paypalGateway
	}

	stripeConfig, err := configService.Get("ecommerce.stripe")
	{
		secret, err := stripeConfig.String("secret")

		if err != nil {
			panic("Could not get config data to initialize stripe client. (secret)")
		}

		stripeClient := &client.API{}
		stripeClient.Init(secret, nil)
		stripeGateway := &payments.Stripe{
			Client: stripeClient,
		}

		paymentGateways["stripe"] = stripeGateway
	}

	paymentGateways["offline"] = &payments.Offline{}

	p := payments.GetModule(paymentGateways)

	// Provide graph with service instances
	err = g.Provide(
		&inject.Object{Value: log, Complete: true},
		&inject.Object{Value: configService, Complete: true},
		&inject.Object{Value: mongoService, Complete: true},
		&inject.Object{Value: errorService, Complete: true},
		&inject.Object{Value: cacheService, Complete: true},
		&inject.Object{Value: s3Service, Complete: true},
		&inject.Object{Value: s3BucketService, Complete: true},
		&inject.Object{Value: firebaseService, Complete: true},
		&inject.Object{Value: statsService, Complete: true},
		&inject.Object{Value: searchService, Complete: true},
		&inject.Object{Value: transmitService, Complete: true},
		&inject.Object{Value: aclService, Complete: false},
		&inject.Object{Value: storeService, Complete: false},
		&inject.Object{Value: assetsService, Complete: false},
		&inject.Object{Value: userService, Complete: false},
		&inject.Object{Value: componentsService, Complete: false},
		&inject.Object{Value: gamingService, Complete: false},
		&inject.Object{Value: mailService, Complete: false},
		&inject.Object{Value: p, Complete: false},
		&inject.Object{Value: broadcaster, Complete: true, Name: "Notifications"},
		&inject.Object{Value: &cliModule},
		&inject.Object{Value: &queueModule},
		&inject.Object{Value: &securityModule},
		&inject.Object{Value: &notificationsModule},
		&inject.Object{Value: &feedModule},
		&inject.Object{Value: &transmitModule},
		&inject.Object{Value: &gcommerceModule},
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

	var cmdPreprocessor = &cobra.Command{
		Use:   "pre-processor",
		Short: "Starts Pre-Processor",
		Long: `Starts API web server listening
        in the specified env port
        `,
		Run: func(cmd *cobra.Command, args []string) {

			// Populate dependencies using the already instantiated DI
			preprocessor.Populate(g)

			// Run preprocessor module
			preprocessor.Run()
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

	var cmdWorkerRoutine = &cobra.Command{
		Use:   "worker [queue]",
		Short: "Starts worker for certain queue",
		Long: `Starts a worker daemon to
		proccess jobs from certain IronMQ queue`,
		Run: func(cmd *cobra.Command, args []string) {

			// Populate dependencies using the already instantiated DI
			queueModule.Populate(g)
			queueModule.Listen(args[0])
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

	var cmdRunTransmit = &cobra.Command{
		Use:   "transmit",
		Short: "Run transmit server",
		Long: `Run transmit server, which
		includes a socket-io instance and a zeromq pull server
        `,
		Run: func(cmd *cobra.Command, args []string) {

			// Populate the DI with the instances
			if err := g.Populate(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			transmitModule.Run()
		},
	}

	var rootCmd = &cobra.Command{Use: "blacker"}
	rootCmd.AddCommand(cmdApi)
	rootCmd.AddCommand(cmdSyncGamification)
	rootCmd.AddCommand(cmdSyncRanking)
	rootCmd.AddCommand(cmdWorkerRoutine)
	rootCmd.AddCommand(cmdJobs)
	rootCmd.AddCommand(cmdRunRoutine)
	rootCmd.AddCommand(cmdRunTransmit)
	rootCmd.AddCommand(cmdPreprocessor)
	rootCmd.Execute()

	return
}

func string_value(value string, err error) string {

	if err != nil {
		panic(err)
	}

	return value
}
