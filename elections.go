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
    "bytes"
)

type ElectionOption struct {
    UserId  bson.ObjectId `bson:"user_id" json:"user_id"`
    Content string `bson:"content" json:"content"`
    User    interface{} `bson:"author,omitempty" json:"author,omitempty"`
    Votes   Votes `bson:"votes" json:"votes"`
    Created time.Time `bson:"created_at" json:"created_at"`
}

// ByElectionsCreatedAt implements sort.Interface for []Person based on
// the Age field.
type ByElectionsCreatedAt []ElectionOption

func (a ByElectionsCreatedAt) Len() int           { return len(a) }
func (a ByElectionsCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByElectionsCreatedAt) Less(i, j int) bool { return !a[i].Created.Before(a[j].Created) }

func ElectionAddOption (r render.Render, database *mgo.Database, req *http.Request, params martini.Params) {
    
    if bson.IsObjectIdHex(params["id"]) == false {
        
        response := map[string]string{
		    "error":  "Invalid params.",
		    "status": "507",
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
		    "status": "502",
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
		    "status": "503",
	    }
        
        r.JSON(401, response)
        
        return 
	}
	
    // Get the option content
    body, err := ioutil.ReadAll(req.Body)    
    
    if err != nil {
        
        panic(err)   
    }
    
    var option map[string] interface{}
    
    err = json.Unmarshal(body, &option)
    
    if err != nil {
        
        panic(err)   
    }
    
    content, okay := option["content"]
    
    if okay && content != "" {
        
        component, okay := option["component"]
        
        if okay && component != "" {
            
            valid := false
            
            // Validate the component name to be avoid injections
            for _, possible := range avaliable_components {
             
                if component == possible {
                    
                    valid = true   
                }
            }
            
            if valid == false {
                
                response := map[string]string{
        		    "error":  "Invalid request, dont attempt hacking.",
        		    "status": "505",
        	    }
                
                r.JSON(400, response)
                
                return
            }
            
            // Get the post using the slug
            id := bson.ObjectIdHex(params["id"])
        
            // Posts collection
            collection := database.C("posts")
            
            var post Post
            
            err := collection.FindId(id).One(&post)    
            
            if err != nil {
                
                response := map[string]string{
        		    "error":  "Couldnt found post with that id.",
        		    "status": "501",
        	    }
                
                r.JSON(404, response)
                
                return
            }
            
            votes := Votes {
                Up: "0",
                Down: "0",
            }
            
            election := ElectionOption {
                UserId: user_token.UserId,
                Content: content.(string),
                Votes: votes,
                Created: time.Now(),
            }
            
            var push bytes.Buffer
            
            // Make the push string
            push.WriteString("components.")
            push.WriteString(component.(string))
            push.WriteString(".options")
            
            over := push.String()
            
            change := bson.M{"$push": bson.M{over: election}, "$set": bson.M{"updated_at": time.Now()}}
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
        	    "status": "506",
        	    "message": "okay",
            }
            
            r.JSON(200, response)
    
            return
        }
    }
    
    response := map[string]string{
	    "error":  "Not authorized",
	    "status": "504",
    }
    
    r.JSON(401, response)
}