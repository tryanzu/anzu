package main

import (
    "net/http"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"
    "time"
    "strings"
)

type Playlist struct {
    Id bson.ObjectId `bson:"_id" json:"id"`
    Name string `bson:"name" json:"name"`   
    Description  string `bson:"description" json:"description"`   
    Use       string `bson:"use" json:"use"`  
    Active    bool `bson:"active" json:"active"`  
    External  string `bson:"external" json:"external"`
    Created time.Time `bson:"created_at" json:"created_at"`
    Updated time.Time `bson:"updated_at" json:"updated_at"`
}

func PlaylistGetList (r render.Render, database *mgo.Database, req *http.Request, params martini.Params) {
    
    sections := params["sections"]
    
    // List of sections to get
    list := strings.Split(sections, ",")
    
    // Get the collection
    collection := database.C("playlists")
    
    var result []Playlist
    
    err := collection.Find(bson.M {"use": bson.M {"$in": list}}).All(&result)
    
    if err != nil {
        
        panic(err)   
    }
    
    if len(result) == 0 {
        
        // No results
        r.JSON(200, []string{})
        
        return    
    }
    
    r.JSON(200, result)
}