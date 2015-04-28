package main

import (
    //"github.com/robfig/cron"
    "gopkg.in/mgo.v2"
    "fmt"
)

func main() {
    
    fmt.Println("Hammer starting...")
    
    // Start a session with out replica set
    session, err := mgo.Dial("mongodb://alpha:3De8ctNNQK0y@@capital.3.mongolayer.com:10069,capital.2.mongolayer.com:10066/nagios")
    database := session.DB("nagios")
    
    if err != nil {
        panic(err)
    }
    
    defer session.Close()
    
    // Optional. Switch the session to a monotonic behavior.
    session.SetMode(mgo.Monotonic, true)
    
    CalculateUsersStats(database)
    
    // return
    
    // c := cron.New()
    
    // // Some jobs to be executed every 2 hours
    // c.AddFunc("@every 1h", func() { 
        
    //     CalculateUsersStats(database)
    // })
    
    // // Start the jobs :D
    // c.Start()
    
    select {}
}