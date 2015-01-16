package main

import (
    "gopkg.in/mgo.v2/bson"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
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

type ElectionForm struct {
    Component  string `json:"component" binding:"required"`     
    Content  string `json:"content" binding:"required"`  
}

// ByElectionsCreatedAt implements sort.Interface for []ElectionOption based on Created field
type ByElectionsCreatedAt []ElectionOption

func (a ByElectionsCreatedAt) Len() int           { return len(a) }
func (a ByElectionsCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByElectionsCreatedAt) Less(i, j int) bool { return !a[i].Created.Before(a[j].Created) }

func ElectionAddOption (c *gin.Context) {
    
    id := c.Params.ByName("id")
    
    if bson.IsObjectIdHex(id) == false {
        
        // Invalid request
        c.JSON(400, gin.H{"error": "Invalid request...", "status": 507})
        
        return    
    }
    
    // Get the query parameters
    qs := c.Request.URL.Query()
    
    // Name of the set to get
    token := qs.Get("token")
    
    if token == "" {
        
        c.JSON(401, gin.H{"error": "Not authorized...", "status": 502})
        return
    }
    
    // Get user by token
    user_token  := UserToken{}
    
    // Try to fetch the user using token header though
	err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)
	
	if err != nil {
	 
	    c.JSON(401, gin.H{"error": "Not authorized...", "status": 503})
        
        return 
	}
	
    var option ElectionForm
    
    if c.BindWith(&option, binding.JSON) {
        
        // Check if component is valid
        component := option.Component   
        content := option.Content
        valid := false
        
        for _, possible := range avaliable_components {
         
            if component == possible {
                
                valid = true   
            }
        }
        
        if valid == true {
                
            // Posts collection
            collection := database.C("posts")
            
            var post Post
            
            err := collection.FindId(id).One(&post)    
            
            if err != nil {
                
                // No guest can vote
                c.JSON(404, gin.H{"error": "Couldnt found post with that id.", "status": 501})
                
                return
            }
            
            votes := Votes {
                Up: 0,
                Down: 0,
            }
            
            election := ElectionOption {
                UserId: user_token.UserId,
                Content: content,
                Votes: votes,
                Created: time.Now(),
            }
            
            var push bytes.Buffer
            
            // Make the push string
            push.WriteString("components.")
            push.WriteString(component)
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
	        
	        c.JSON(200, gin.H{"message": "okay", "status": 506})
            return
        }    
    }
    
    c.JSON(401, gin.H{"error": "Not authorized", "status": 504})
}