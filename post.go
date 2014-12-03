package main

import (
    "net/http"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"
    "time"
)

type Votes struct {
    Up int `bson:"up" json:"up"`
    Down int `bson:"down" json:"down"`
    Rating int `bson:"rating,omitempty" json:"rating,omitempty"`
}

type Author struct {
    Id     bson.ObjectId `bson:"id,omitempty" json:"id,omitempty"`
    Name   string `bson:"name" json:"name"`
    Email  string `bson:"email" json:"email"`
    Avatar string `bson:"avatar" json:"avatar"`
}

type Comments struct {
    Count int `bson:"count" json:"count"` 
    Set   []Comment `bson:"set" json:"set"`
}

type Comment struct {
    Author  Author `bson:"author" json:"author"`
    Votes   Votes `bson:"votes" json:"votes"`
    Content string `bson:"content" json:"content"`
    Created time.Time `bson:"created_at" json:"created_at"`
}

type Post struct {
    Id bson.ObjectId `bson:"_id" json:"id"`
    Title string `bson:"title" json:"title"`   
    Slug  string `bson:"slug" json:"slug"`   
    Type  string `bson:"type" json:"type"`   
    Content string `bson:"content" json:"content"`   
    Categories []string `bson:"categories" json:"categories"`   
    Comments   Comments `bson:"comments" json:"comments"`   
    Author     Author `bson:"author" json:"author"`  
    Votes      Votes `bson:"votes" json:"votes"`
    Components interface{} `bson:"components" json:"components"`
    Created time.Time `bson:"created_at" json:"created_at"`
    Updated time.Time `bson:"updated_at" json:"updated_at"`
}

func PostsGet (r render.Render, database *mgo.Database, req *http.Request) {
    
    // Get the query parameters
    qs := req.URL.Query()
    
    // Name of the set to get
    named := qs.Get("named")
    
    var results []Post
    
    // Get the collection
    collection := database.C("posts")
    query := collection.Find(bson.M{}).Limit(10)
    
    if named == "" || named == "default" {
        
        // Sort from newest to oldest
        query = query.Sort("-created_at")
    }
    
    // Try to fetch the posts
	err := query.All(&results)
	
	if err != nil {
	    
	    panic(err)
	}
    
    r.JSON(200, results)
}

func PostsGetOne (r render.Render, database *mgo.Database, req *http.Request, params martini.Params) {
    
    if bson.IsObjectIdHex(params["id"]) == false {
        
        response := map[string]string{
		    "error":  "Invalid params to get a post.",
		    "status": "202",
	    }
        
        r.JSON(400, response)
        
        return   
    }
    
    // Get the id of the needed post
    id := bson.ObjectIdHex(params["id"])
    
    // Get the collection
    collection := database.C("posts")
    
    post := Post{}
    
    // Try to fetch the needed post by id
    err := collection.FindId(id).One(&post)
    
    if err != nil {
        
        response := map[string]string{
		    "error":  "Couldnt found post with that id.",
		    "status": "201",
	    }
        
        r.JSON(404, response)
        
        return
    }
    
    r.JSON(200, post)    
}

func PostsGetOneSlug (r render.Render, database *mgo.Database, req *http.Request, params martini.Params) {
    
    // Get the post using the slug
    slug := params["slug"]
    
    // Get the collection
    collection := database.C("posts")
    
    post := Post{}
    
    // Try to fetch the needed post by id
    err := collection.Find(bson.M{"slug": slug}).One(&post)
    
    if err != nil {
        
        response := map[string]string{
		    "error":  "Couldnt found post with that slug.",
		    "status": "203",
	    }
        
        r.JSON(404, response)
        
        return
    }
    
    r.JSON(200, post)  
}