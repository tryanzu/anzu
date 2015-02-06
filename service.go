package main

import (
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"github.com/fernandez14/deferclient/deferstats"
	"github.com/gin-gonic/gin"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"os"
	"runtime"
	"time"
)

var (
	database *mgo.Database
	mongo    *mgo.Session
	config   gin.H
	redis    *goredis.Redis
	socket   *zmq.Socket
	zmq_messages chan string
)

func main() {

	// Start by using the power of the machine cores
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	// Run with the specified env file
	envfile := os.Getenv("ENV_FILE")

	if envfile == "" {

		envfile = "./env.json"
	}
	
	// Read config from env file
	configFile, err := ioutil.ReadFile(envfile)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	err = json.Unmarshal(configFile, &config)

	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	
    config_services := config["plugins"].(map[string]interface{})
	
	// Start the zmq socket
	socket, err  := zmq.NewSocket(zmq.REQ)
    if err != nil { panic(err) }
    
    socket.Connect(config_services["socket"].(string))
    
    // Channel to send the socket messages
    zmq_messages = make(chan string)
    
    // Deferpanic client
	deferstats.Token = config_services["deferpanic"].(string)
	deferstats.Verbose = false

	// Start gin classic middlewares
	g := gin.Default()

	// Use deferpanic recovery
	//g.Use(recovery.Recovery())

	// Use cors middleware
	g.Use(CORS())

	g.Use(func(c *gin.Context) {

		startTime := time.Now()

		c.Next()

		// Add the request to the list of played urls
		deferstats.AddRequest(startTime, c.Request.URL.Path)
	})

	// Start a session with redis-service
	config_redis := config["cache"].(map[string]interface{})
	redis, err = goredis.Dial(&goredis.DialConfig{Address: config_redis["redis"].(string)})

	if err != nil {
		panic(err)
	}

	// Start a session with replica set
	config_database := config["database"].(map[string]interface{})
	database, mongo = databaseInit(config_database["uri"].(string), config_database["name"].(string))

	// Close the database connection when needed
	defer mongo.Close()

	v1 := g.Group("/v1")
	{
		// Comment routes
		v1.POST("/post/comment/:id", CommentAdd)

		// Post routes
		v1.GET("/feed", FeedGet)
		v1.GET("/post", PostsGet)
		//v1.GET("/post/:id", PostsGetOne)
		v1.GET("/post/s/:slug", PostsGetOneSlug)
		v1.POST("/post", PostCreate)

		// // Election routes
		v1.POST("/election/:id", ElectionAddOption)

		// // Votes routes
		v1.POST("/vote/comment/:id", VoteComment)
		v1.POST("/vote/component/:id", VoteComponent)

		// User routes
		v1.POST("/user", UserRegisterAction)
		v1.GET("/user/my/notifications", UserNotificationsGet)
		v1.GET("/user/my", UserGetByToken)
		v1.GET("/user/activity", UserInvolvedFeedGet)
		v1.GET("/user/search", UserAutocompleteGet)
		v1.PUT("/user/my", UserUpdateProfile)
		v1.POST("/user/get-token/facebook", UserGetTokenFacebook)
		v1.GET("/user/get-token", UserGetToken)
		// v1.GET("/user/:id", UserGetOne)
		
		// Messaging routes
		v1.GET("/messages", MessagesGet)
		v1.POST("/messages", MessagePublish)
		v1.GET("/hashtags", HashtagsGet)

		// Playlist routes
		v1.GET("/playlist/l/:sections", PlaylistGetList)

		// Categories routes
		v1.GET("/category", CategoriesGet)
	}

	// Run over the 3000 port
	port := os.Getenv("RUN_OVER")

	if port == "" {

		port = "3000"
	}

	// Collect stats
	go deferstats.CaptureStats()
	
	// Wait for zmq messages to send to the socket server
	go func(zmq_messages chan string) {
    	
    	for {
    		select {
    		case message := <- zmq_messages:	
	    		
	    		// Send to the socket
	    		socket.Send(message, 0)
	    		socket.Recv(0)
    		}
    	}
    }(zmq_messages)
    
	g.Run(":" + port)
}

func databaseInit(connection string, name string) (*mgo.Database, *mgo.Session) {

	// Start a session with out replica set
	session, err := mgo.Dial(connection)

	if err != nil {

		// There has been an error connection to the database
		panic(err)
	}

	database := session.DB(name)

	// Set monotonic session behavior
	session.SetMode(mgo.Monotonic, true)

	return database, session
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Auth-Token,Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			fmt.Println("options")
			c.Abort(200)
			return
		}
		c.Next()
	}
}