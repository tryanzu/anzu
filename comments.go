package main 

import (
    "net/http"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"
    "io/ioutil"
    "encoding/json"
    "time"
)

func CommentAdd (r render.Render, database *mgo.Database, req *http.Request, params martini.Params) { 
    
    if bson.IsObjectIdHex(params["id"]) == false {
        
        response := map[string]string{
		    "error":  "Invalid params.",
		    "status": "701",
	    }
        
        r.JSON(400, response)
        return   
    }   
    
    // Get the query parameters
    qs := req.URL.Query()
    
    // Name of the set to get
    token := qs.Get("token")
    
    if token == "" {
        
        response := map[string]string{
		    "error":  "Not authorized",
		    "status": "702",
	    }
        
        r.JSON(401, response)
        
        return
    }
    
    // Get user by token
    user_token  := UserToken{}
    
    // Try to fetch the user using token header though
	err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)
	
	if err != nil {
	    response := map[string]string{
		    "error":  "Not authorized",
		    "status": "703",
	    }
        r.JSON(401, response)
        return 
	}
	
	// Get the option content
    body, err := ioutil.ReadAll(req.Body)    
    
    if err != nil {
        
        panic(err)   
    }
    
    var comment map[string] interface{}
    
    err = json.Unmarshal(body, &comment)
    
    if err != nil {
        
        panic(err)   
    }
    
    content, okay := comment["content"]
    
    if okay && content != "" {
        
        // Get the post using the slug
        id := bson.ObjectIdHex(params["id"])
    
        // Posts collection
        collection := database.C("posts")
        
        var post Post
        
        err := collection.FindId(id).One(&post)    
        
        if err != nil {
            
            response := map[string]string{
    		    "error":  "Couldnt found post with that id.",
    		    "status": "705",
    	    }
            
            r.JSON(404, response)
            
            return
        }
        
        votes := Votes{
            Up: 0,
            Down: 0,
        }
        
        comment := Comment{
            UserId: user_token.UserId,
            Votes: votes,
            Content: content.(string),
            Created: time.Now(),
        }
        
        // Update the post and push the comments
        change := bson.M{"$push": bson.M{"comments.set": comment}, "$set": bson.M{"updated_at": time.Now()}, "$inc": bson.M{"comments.count": 1}}
        err = collection.Update(bson.M{"_id": post.Id}, change)
        
        if err != nil {
    		panic(err)
    	}
    
        // Check if we need to add participant
        users := post.Users
        need_add := true
        
        for _, already_within := range users {
            
            if already_within == user_token.UserId {
                
                need_add = false   
            }
        }
        
        if need_add == true {
	            
            // Add the user to the user list
            change := bson.M{"$push": bson.M{"users": user_token.UserId}}
            err = collection.Update(bson.M{"_id": post.Id}, change)   
            
            if err != nil {
    		    panic(err)
    	    }
        }
        
        response := map[string]string{
    	    "status": "706",
    	    "message": "okay",
        }
        
        r.JSON(200, response)

        return
    }
    
    response := map[string]string{
	    "error":  "Not authorized",
	    "status": "704",
    }
    
    r.JSON(401, response)
}