package main

import (
	"fmt"
    "os"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "github.com/fernandez14/turnpike"
    "github.com/gin-gonic/gin"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    zmq "github.com/pebbe/zmq4"
)

var (
    database *mgo.Database
	mongo    *mgo.Session
    config   gin.H
)

type Topic struct {
    Name    string `bson:"name"`
    Session string `bson:"session"`
    Status  bool   `bson:"status"`
}

func listenUpZeroMq(server *turnpike.Server) {
    
	ports := config["ports"].(map[string]interface{})
    fmt.Println("Listening for incoming zmq messages on port " + ports["zmq"].(string))
    
    responder, _ := zmq.NewSocket(zmq.REP)
    defer responder.Close()
	responder.Bind("tcp://*:" + ports["zmq"].(string))
	
	for {
		//  Wait for next request from client
		msg, _ := responder.Recv(0)
		
        // Unmarshal the gotten json(hopefully valid)
        var data interface{}
        
        if err := json.Unmarshal([]byte(msg), &data); err != nil {
            
            fmt.Println("Got invalid message", msg)
            
            continue
        }
        
        // Data cast data since interface{} itself doesnt support indexing by keys
        dataCast := data.(map[string] interface{})
        
        if topic, check := dataCast["to"]; check {
             
            fmt.Printf("Got %v to %s \n", string(msg), topic.(string))
            
            // We have to send the event to topic(topic var has been type asserted)
            server.SendEvent(topic.(string), data)
        }
        
		responder.Send("okay", 0)
	}
}

func main() {
    
    // Create new wamp server with turnpike
    server := turnpike.NewServer(true)
    
    // Run with the specified env file
	envfile := os.Getenv("ENV_FILE")

	if envfile == "" {

		envfile = "../env.json"
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
	
	// Start a session with replica set
	config_database := config["database"].(map[string]interface{})
	database, mongo = databaseInit(config_database["uri"].(string), config_database["name"].(string))
    
    // Close the database connection when needed
	defer mongo.Close()
	
    server.SetSessionClosedCallback(subhandlerClosed)
    
    server.RegisterSubHandler("", subhandlerOpened)
    
    // Listen up for incoming messages to be broadcasted through zeromq (channeled)
    go listenUpZeroMq(server)
    
	// Ports config
	ports := config["ports"].(map[string]interface{})
	
    // Handle the websocket server
    http.Handle("/", server.Handler)
    fmt.Println("Listening on port " + ports["socket"].(string))
	
    if err := http.ListenAndServe(":" + ports["socket"].(string), nil); err != nil {
        
        fmt.Println("Error: ")
        fmt.Printf("%v", err)
    }
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

func subhandlerOpened(client string, topic string) bool {
        
    t := &Topic{
        Name: topic,
        Session: client,
        Status: true,
    }
    
    // Insert without panicking 
    database.C("socket").Insert(t)
    
    return true
}

func subhandlerClosed(client string) {
        
    // Remove the session
    database.C("socket").Update(bson.M{"session": client}, bson.M{"$set": bson.M{"status": false}})
}