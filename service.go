package main

import (
    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"
    "gopkg.in/mgo.v2"
)

func main() {
    
    // Start martini classic
    m := martini.Classic()
    
    // Use the render from the martini-contrib
    m.Use(render.Renderer())
    
    // Start a session with out replica set
    session, err := mgo.Dial("mongodb://alpha:3De8ctNNQK0y@@capital.3.mongolayer.com:10069,capital.2.mongolayer.com:10066/nagios")
    database := session.DB("nagios")
    
    if err != nil {
        panic(err)
    }
    
    defer session.Close()
        
    // Optional. Switch the session to a monotonic behavior.
    session.SetMode(mgo.Monotonic, true)
    
    m.Get("/", func () string {
        
        return "Welcome to the blacker API"
    })
    
    m.Group("/v1", func(r martini.Router) {
        
        // API routes
        r.Get("/post", PostsGet)
        r.Get("/post/s/(?P<slug>[a-zA-Z0-9-_.]+)", PostsGetOneSlug)
        r.Get("/post/:id", PostsGetOne)
        r.Get("/user/my", UserGetByToken)
        r.Get("/user/get-token", UserGetToken)
        r.Get("/user/:id", UserGetOne)
        r.Get("/playlist/l/:sections", PlaylistGetList)
    })
    
    // Map database service to whole API
    m.Map(database)
    m.Run()
}