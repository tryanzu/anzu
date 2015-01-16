package main

import (
    "gopkg.in/mgo.v2/bson"
    "github.com/gin-gonic/gin"
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

func PlaylistGetList (c *gin.Context) {
    
    sections := c.Params.ByName("sections")
    
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
        c.JSON(200, []string{})
        
        return    
    }
    
    c.JSON(200, result)
}