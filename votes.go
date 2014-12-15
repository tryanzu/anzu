package main

import (
    "net/http"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"
    "bytes"
    "time"
    "io/ioutil"
    "encoding/json"
)

type Vote struct {
    Id bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
    UserId     bson.ObjectId `bson:"user_id" json:"user_id"`
    Type       string `bson:"type" json:"type"`
    NestedType string `bson:"nested_type" json:"nested_type"`
    RelatedId  bson.ObjectId `bson:"related_id" json:"related_id"`
    Value      int `bson:"value" json:"value"`
    Created time.Time `bson:"created_at" json:"created_at"`
}

func VoteComponent (r render.Render, database *mgo.Database, req *http.Request, params martini.Params) {
    
    if bson.IsObjectIdHex(params["id"]) == false {
        
        response := map[string]string{
		    "error":  "Invalid params.",
		    "status": "601",
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
		    "status": "602",
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
		    "status": "603",
	    }
        
        r.JSON(401, response)
        
        return 
	}
	
    // Get the option content
    body, err := ioutil.ReadAll(req.Body)    
    
    if err != nil {
        
        panic(err)   
    }
    
    var vote map[string] interface{}
    
    err = json.Unmarshal(body, &vote)
    
    if err != nil {
        
        panic(err)   
    }
    
    direction, okay := vote["direction"]
    
    if okay && (direction == "up" || direction == "down") {
        
        component, okay := vote["component"]
        
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
        		    "status": "604",
        	    }
                
                r.JSON(400, response)
                
                return
            }
            
            // Add the vote itself to the votes collection
            var value int
            
            if direction == "up" {
                
                value = 1;
            }
            
            if direction == "down" {
                
                value = -1;   
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
        		    "status": "605",
        	    }
                
                r.JSON(404, response)
                
                return
            }
            
            var already_voted Vote
            
            err = database.C("votes").Find(bson.M{"type": "component", "nested_type": component.(string), "user_id": user_token.UserId, "related_id": id}).One(&already_voted)               
            
            var add bytes.Buffer
            
            // Make the push string
            add.WriteString("components.")
            add.WriteString(component.(string))
            add.WriteString(".votes.")
            add.WriteString(direction.(string))
            
            inc := add.String()
            
            if err == nil {
                
                change := bson.M{"$inc": bson.M{inc: -1}}
                err = collection.Update(bson.M{"_id": post.Id}, change)
                
                if err != nil {
                    
                    panic(err)   
                }
                
                err = database.C("votes").RemoveId(already_voted.Id)
                
                if err != nil {
                    
                    panic(err)
                }
                
                response := map[string]string{
            	    "status": "609",
            	    "message": "okay",
                }
                
                r.JSON(200, response)
        
                return
            }
            
            change := bson.M{"$inc": bson.M{inc: 1}}
            err = collection.Update(bson.M{"_id": post.Id}, change)
            
            if err != nil {
                
                panic(err)   
            }
            
            vote := &Vote{
                UserId: user_token.UserId,
                Type: "component",
                NestedType: component.(string),
                RelatedId: id,
                Value: value,
                Created: time.Now(),
            }
            
            err = database.C("votes").Insert(vote)
            
            response := map[string]string{
        	    "status": "606",
        	    "message": "okay",
            }
            
            r.JSON(200, response)
    
            return
        }
    }
    
    response := map[string]string{
	    "error":  "Not authorized",
	    "status": "608",
    }
    
    r.JSON(401, response)
}