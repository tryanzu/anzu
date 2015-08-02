package main

import (
	"fmt"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/cosn/firebase"
	"github.com/facebookgo/inject"
	"github.com/fernandez14/spartangeek-blacker/handle"
	"github.com/fernandez14/spartangeek-blacker/modules/notifications"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/robfig/cron"
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
	var posts handle.PostAPI
	var votes handle.VoteAPI
	var users handle.UserAPI
	var categories handle.CategoryAPI
	var elections handle.ElectionAPI
	var comments handle.CommentAPI
	var parts handle.PartAPI
	var stats handle.StatAPI
	var middlewares handle.MiddlewareAPI
	var collector handle.CollectorAPI
	var sitemap handle.SitemapAPI
	var errors handle.ErrorAPI
	var notifications notifications.NotificationsModule

	// Services for the DI
	configService, _ := config.ParseJsonFile(envfile)
	mongoService := mongo.NewService(string_value(configService.String("database.uri")), string_value(configService.String("database.name")))
	errorService, _ := raven.NewClient(string_value(configService.String("sentry.dns")), nil)
	cacheService, _ := goredis.Dial(&goredis.DialConfig{Address: string_value(configService.String("cache.redis"))})
	firebaseService := new(firebase.Client)
	firebaseService.Init(string_value(configService.String("firebase.url")), string_value(configService.String("firebase.secret")), nil)
	gamingService := new(handle.GamingAPI)

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
		&inject.Object{Value: configService, Complete: true},
		&inject.Object{Value: mongoService, Complete: true},
		&inject.Object{Value: errorService, Complete: true},
		&inject.Object{Value: cacheService, Complete: true},
		&inject.Object{Value: s3Service, Complete: true},
		&inject.Object{Value: s3BucketService, Complete: true},
		&inject.Object{Value: firebaseService, Complete: true},
		&inject.Object{Value: statsService, Complete: true},
		&inject.Object{Value: gamingService, Complete: false},
		&inject.Object{Value: &notifications},
		&inject.Object{Value: &collector},
		&inject.Object{Value: &errors},
		&inject.Object{Value: &posts},
		&inject.Object{Value: &votes},
		&inject.Object{Value: &users},
		&inject.Object{Value: &categories},
		&inject.Object{Value: &elections},
		&inject.Object{Value: &comments},
		&inject.Object{Value: &parts},
		&inject.Object{Value: &stats},
		&inject.Object{Value: &middlewares},
		&inject.Object{Value: &sitemap},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Populate the DI with the instances
	if err := g.Populate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// After DI populate initializations
	gamingService.Init()

	// Start gin classic middlewares
	router := gin.Default()

	router.GET("/loaderio-9d2a383a6a8ed580a4a26cf61ce183c3/", func(c *gin.Context) {

		c.String(200, "loaderio-9d2a383a6a8ed580a4a26cf61ce183c3")
	})
	// Middlewares setup
	router.Use(middlewares.ErrorTracking())
	router.Use(middlewares.CORS())
	router.Use(middlewares.MongoRefresher())
	router.Use(middlewares.StatsdTiming())

	// Sitemap generator
	router.GET("/sitemap.xml", sitemap.GetSitemap)

	v1 := router.Group("/v1")

	v1.Use(middlewares.Authorization())
	{
		v1.POST("/subscribe", users.UserSubscribe)
		
		// Gamification routes
		v1.GET("/gamification", gamingService.GetRules)
		
		// Post routes
		v1.GET("/feed", posts.FeedGet)
		v1.GET("/post", posts.FeedGet)
		v1.GET("/posts/:id", posts.PostsGetOne)
		v1.GET("/post/s/:id", posts.PostsGetOne)

		// // Election routes
		v1.POST("/election/:id", elections.ElectionAddOption)

		// User routes
		v1.POST("/user", users.UserRegisterAction)
		//v1.GET("/user/my/notifications", users.UserNotificationsGet)
		v1.GET("/users/:id", users.UserGetOne)
		v1.GET("/user/activity", users.UserInvolvedFeedGet)
		v1.GET("/user/search", users.UserAutocompleteGet)
		v1.POST("/user/get-token/facebook", users.UserGetTokenFacebook)
		v1.GET("/user/get-token", users.UserGetToken)
		v1.GET("/auth/get-token", users.UserGetJwtToken)

		// Messaging routes
		//v1.GET("/messages", MessagesGet)
		//v1.POST("/messages", MessagePublish)
		//v1.GET("/hashtags", HashtagsGet)

		// Playlist routes
		//v1.GET("/playlist/l/:sections", PlaylistGetList)

		// Categories routes
		v1.GET("/category", categories.CategoriesGet)

		// Parts routes
		v1.GET("/part", parts.GetPartTypes)
		v1.GET("/part/:type/manufacturers", parts.GetPartManufacturers)
		v1.GET("/part/:type/models", parts.GetPartManufacturerModels)

		// Stats routes
		v1.GET("/stats/board", stats.BoardGet)

		authorized := v1.Group("")

		authorized.Use(middlewares.NeedAuthorization())
		{
			// Comment routes
			authorized.POST("/post/comment/:id", comments.CommentAdd)

			// Post routes
			authorized.POST("/post", posts.PostCreate)

			// User routes
			v1.POST("/user/my/avatar", users.UserUpdateProfileAvatar)
			v1.GET("/user/my", users.UserGetByToken)
			v1.PUT("/user/my", users.UserUpdateProfile)

			// // Votes routes
			v1.POST("/vote/comment/:id", votes.VoteComment)
			v1.POST("/vote/component/:id", votes.VoteComponent)
			v1.POST("/vote/post/:id", votes.VotePost)
		}
	}

	// Run over the 3000 port
	port := os.Getenv("RUN_OVER")
	if port == "" {
		port = "3000"
	}

	// Start the jobs
	c := cron.New()

	// Reset the user temporal stuff each X
	c.AddFunc("@midnight", gamingService.ResetTempStuff)

	// Start the jobs
	c.Start()

	router.Run(":" + port)
}

func string_value(value string, err error) string {

	if err != nil {
		panic(err)
	}

	return value
}
