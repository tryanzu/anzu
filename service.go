package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/getsentry/raven-go"
	"github.com/facebookgo/inject"
	"github.com/olebedev/config"
	"github.com/xuyu/goredis"	
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/fernandez14/spartangeek-blacker/handle"
	"os"
	//"errors"
	"runtime"
)

var (
	zmq_messages chan string
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
	
	// Services for the DI
	configService, _ := config.ParseJsonFile(envfile)
	mongoService     := mongo.NewService(string_value(configService.String("database.uri")), string_value(configService.String("database.name")))
	errorService, _  := raven.NewClient(string_value(configService.String("sentry.dns")), nil)
	cacheService, _  := goredis.Dial(&goredis.DialConfig{Address: string_value(configService.String("cache.redis"))})

	// Provide graph with service instances
	err := g.Provide(
        &inject.Object{Value: configService},
        &inject.Object{Value: mongoService},
        &inject.Object{Value: errorService},
        &inject.Object{Value: cacheService},
        &inject.Object{Value: &posts},
        &inject.Object{Value: &votes},
        &inject.Object{Value: &users},
        &inject.Object{Value: &categories},
        &inject.Object{Value: &elections},
        &inject.Object{Value: &comments},
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
	
    // Channel to send the socket messages
    zmq_messages = make(chan string)
    
	// Start gin classic middlewares
	router := gin.Default()

	// Exception tracking setup
	//g.Use(DeferTracking())
	router.Use(CORS())

	v1 := router.Group("/v1")
	{
		// Comment routes
		v1.POST("/post/comment/:id", comments.CommentAdd)

		// Post routes
		v1.GET("/feed", posts.FeedGet)
		v1.GET("/post", posts.PostsGet)
		v1.GET("/posts/:id", posts.PostsGetOne)
		v1.GET("/post/s/:slug", posts.PostsGetOneSlug)
		v1.POST("/post", posts.PostCreate)

		// // Election routes
		v1.POST("/election/:id", elections.ElectionAddOption)

		// // Votes routes
		v1.POST("/vote/comment/:id", votes.VoteComment)
		v1.POST("/vote/component/:id", votes.VoteComponent)

		// User routes
		v1.POST("/user", users.UserRegisterAction)
		//v1.GET("/user/my/notifications", users.UserNotificationsGet)
		v1.GET("/user/my", users.UserGetByToken)
		v1.GET("/user/activity", users.UserInvolvedFeedGet)
		v1.GET("/user/search", users.UserAutocompleteGet)
		v1.PUT("/user/my", users.UserUpdateProfile)
		v1.POST("/user/get-token/facebook", users.UserGetTokenFacebook)
		v1.GET("/user/get-token", users.UserGetToken)
		//v1.GET("/user/:id", users.UserGetOne)
		
		// Messaging routes
		//v1.GET("/messages", MessagesGet)
		//v1.POST("/messages", MessagePublish)
		//v1.GET("/hashtags", HashtagsGet)

		// Playlist routes
		//v1.GET("/playlist/l/:sections", PlaylistGetList)

		// Categories routes
		v1.GET("/category", categories.CategoriesGet)

		v1.GET("/error", func(c *gin.Context) {

			panic("Oh no! shit happened")
		})
	}

	// Run over the 3000 port
	port := os.Getenv("RUN_OVER")
	if port == "" {
		port = "3000"
	}
	
	// Wait for zmq messages to send to the socket server
	go func(zmq_messages chan string) {
    	
    	for {
    		select {
    		case _ = <- zmq_messages:	
	    		
	    		// Send to the socket
	    		//socket.Send(message, 0)
	    		//socket.Recv(0)
    		}
    	}
    }(zmq_messages)
    
	router.Run(":" + port)
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Auth-Token,Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {

			c.AbortWithStatus(200)
			return
		}
		c.Next()
	}
}

func string_value(value string, err error) string {

	if err != nil {
		panic(err)
	}

	return value
}

/*func DeferTracking() gin.HandlerFunc {
	return func(c *gin.Context) {

		envfile := os.Getenv("ENV_FILE")

		if envfile == "" {

			envfile = "./env.json"
		}

		tags := map[string]string {
			"config_file": envfile,
		}

		defer func() {

			var packet *raven.Packet

			switch rval := recover().(type) {
			case nil:
				return
			case error:
				packet = raven.NewPacket(rval.Error(), raven.NewException(rval, raven.NewStacktrace(2, 3, nil)))
			default:
				rvalStr := fmt.Sprint(rval)
				packet = raven.NewPacket(rvalStr, raven.NewException(errors.New(rvalStr), raven.NewStacktrace(2, 3, nil)))
			}

			// Grab the error and send it to sentry
			error_tracking.Capture(packet, tags)

			// Also abort the request with 500
			c.AbortWithStatus(500)
		}()

		c.Next()
	}
}*/