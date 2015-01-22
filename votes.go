package main

import (
    "gopkg.in/mgo.v2/bson"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "github.com/ftrvxmtrx/gravatar"
    "fmt"
    "bytes"
    "time"
    "strconv"
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

type VoteForm struct {
    Component  string `json:"component" binding:"required"`
    Direction  string `json:"direction" binding:"required"`
}

type VoteCommentForm struct {
    Comment string `json:"comment" binding:"required"`
}

func VoteComponent (c *gin.Context) {
    
    id := c.Params.ByName("id")
    
    if bson.IsObjectIdHex(id) == false {
        
        // Invalid request
        c.JSON(400, gin.H{"error": "Invalid request...", "status": 601})
        
        return
    }
    
    // Get the query parameters
    qs := c.Request.URL.Query()
    
    // Name of the set to get
    token := qs.Get("token")
    
    // Get user by token
    user_token  := UserToken{}
    
    if token == "" {
        
        // No guest can vote
        c.JSON(401, gin.H{"error": "Not authorized...", "status": 602})
        
        return
        
    } else {
        
        // Try to fetch the user using token header though
    	err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)
    	
    	if err != nil {
    	 
    	    // No guest can vote
            c.JSON(401, gin.H{"error": "Not authorized...", "status": 603})
            
            return
    	}
    }
	
    // Get the vote content
    var vote VoteForm
    
    if c.BindWith(&vote, binding.JSON) {
        
        if vote.Direction == "up" || vote.Direction == "down" {
            
            // Check if component is valid
            component := vote.Component   
            direction := vote.Direction
            valid := false
            
            for _, possible := range avaliable_components {
             
                if component == possible {
                    
                    valid = true   
                }
            }
            
            if valid == true {
                
                // Add the vote itself to the votes collection
                var value int
                
                if direction == "up" {
                    value = 1;
                }
                
                if direction == "down" {
                    value = -1;   
                }
                
                // Get the post using the slug
                id := bson.ObjectIdHex(id)
                collection := database.C("posts")
                
                var post Post
                err := collection.FindId(id).One(&post)    
                
                if err != nil {
                    
                    // No guest can vote
                    c.JSON(404, gin.H{"error": "No post found...", "status": 605})
                    
                    return
                } else {
                    
                    var add bytes.Buffer
                    
                    // Make the push string
                    add.WriteString("components.")
                    add.WriteString(component)
                    add.WriteString(".votes.")
                    add.WriteString(direction)
                    
                    inc := add.String()
            
                    var already_voted Vote
                    
                    err = database.C("votes").Find(bson.M{"type": "component", "user_id": user_token.UserId, "related_id": id, "nested_type": component}).One(&already_voted)               
                    
                    if err == nil {
                        
                        var rem bytes.Buffer
                    
                        // Make the push string
                        rem.WriteString("components.")
                        rem.WriteString(component)
                        rem.WriteString(".votes.")
                        
                        if (direction == "up" && already_voted.Value == 1) || (direction == "down" && already_voted.Value == -1) {
                            
                            rem.WriteString(direction)
                            ctc    := rem.String()
                            change := bson.M{"$inc": bson.M{ctc: -1}}
                            err     = collection.Update(bson.M{"_id": post.Id}, change)
                            
                            if err != nil {
                            
                                panic(err)   
                            }
                            
                            err = database.C("votes").RemoveId(already_voted.Id)
                            
                            if err != nil {
                                
                                panic(err)
                            }
                            
                            c.JSON(200, gin.H{"message": "okay", "status": 609})
                            return
                        
                        } else if (direction == "up" && already_voted.Value == -1) || (direction == "down" && already_voted.Value == 1) {
                            
                            if direction == "up" {
                                   
                                rem.WriteString("down")
                                
                            } else if direction == "down" {
                                
                                rem.WriteString("up")
                            }
                            
                            ctc := rem.String()
                            
                            change := bson.M{"$inc": bson.M{ctc: -1}}
                            
                            err = collection.Update(bson.M{"_id": post.Id}, change)
                            
                            if err != nil {
                            
                                panic(err)   
                            }
                            
                            err = database.C("votes").RemoveId(already_voted.Id)
                            
                            if err != nil {
                                
                                panic(err)
                            }
                        }
                    }
                    
                    change := bson.M{"$inc": bson.M{inc: 1}}
                    err     = collection.Update(bson.M{"_id": post.Id}, change)
                    
                    if err != nil {
                        
                        panic(err)   
                    }
                    
                    vote := &Vote{
                        UserId: user_token.UserId,
                        Type: "component",
                        NestedType: component,
                        RelatedId: id,
                        Value: value,
                        Created: time.Now(),
                    }
                    err = database.C("votes").Insert(vote)
                    
                    c.JSON(200, gin.H{"message": "okay", "status": 606})
                    return
                }
                
            } else {
                
                // Component does not exists
                c.JSON(400, gin.H{"error": "Not authorized...", "status": 604})
            }
        }
    }
    
    c.JSON(401, gin.H{"error": "Couldnt create post, missing information...", "status": 205})
}

func VoteComment (c *gin.Context) {
    
    id := c.Params.ByName("id")
    
    if bson.IsObjectIdHex(id) == false {
        
        // Invalid request
        c.JSON(400, gin.H{"error": "Invalid request...", "status": 601})
        
        return   
    }
    
    // Get the query parameters
    qs := c.Request.URL.Query()
    
    // Name of the set to get
    token := qs.Get("token")
    
    // Get user by token
    user_token  := UserToken{}
    
    if token == "" {
        
        // No guest can vote
        c.JSON(401, gin.H{"error": "Not authorized...", "status": 602})
        
        return
        
    } else {
        
        // Try to fetch the user using token header though
    	err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)
    	
    	if err != nil {
    	 
    	    // No guest can vote
            c.JSON(401, gin.H{"error": "Not authorized...", "status": 603})
            
            return
    	}
    }
	
    // Get the vote content
    var vote VoteCommentForm
    
    if c.BindWith(&vote, binding.JSON) {
        
        // Get the post using the slug
        id := bson.ObjectIdHex(id)
        collection := database.C("posts")
        
        var post Post
        err := collection.FindId(id).One(&post)    
        
        if err != nil {
            
            // No guest can vote
            c.JSON(404, gin.H{"error": "No post found...", "status": 605})
            
            return
        } else {
            
            index := vote.Comment
            
            if _, err := strconv.Atoi(index); err == nil {
                
                var add bytes.Buffer
                
                // Make the push string
                add.WriteString("comments.set.")
                add.WriteString(index)
                add.WriteString(".votes.up")
                
                inc := add.String()
        
                var already_voted Vote
                
                err = database.C("votes").Find(bson.M{"type": "comment", "user_id": user_token.UserId, "related_id": id, "nested_type": index}).One(&already_voted)               
                
                if err == nil {
                    
                    var rem bytes.Buffer
                
                    // Make the push string
                    rem.WriteString("comments.set.")
                    rem.WriteString(index)
                    rem.WriteString(".votes.up")
                    ctc := rem.String()
                    
                    change := bson.M{"$inc": bson.M{ctc: -1}}
                    err     = collection.Update(bson.M{"_id": post.Id}, change)
                    
                    if err != nil {
                    
                        panic(err)   
                    }
                    
                    err = database.C("votes").RemoveId(already_voted.Id)
                    
                    if err != nil {
                        
                        panic(err)
                    }
                    
                    c.JSON(200, gin.H{"message": "okay", "status": 609})
                    return
                }
                
                change := bson.M{"$inc": bson.M{inc: 1}}
                err     = collection.Update(bson.M{"_id": post.Id}, change)
                
                if err != nil {
                    
                    panic(err)   
                }
                
                vote := &Vote{
                    UserId: user_token.UserId,
                    Type: "comment",
                    NestedType: index,
                    RelatedId: id,
                    Value: 1,
                    Created: time.Now(),
                }
                err = database.C("votes").Insert(vote)
                
                comment_index, _ := strconv.Atoi(index)
                comment_ := post.Comments.Set[comment_index]
                
                // Notify the author of the comment
                go func(comment Comment, token UserToken, post Post) {
                    
                    user_id := comment.UserId
                    
                    // Get the comment like author
                    var user User
                
                    err := database.C("users").Find(bson.M{"_id": token.UserId}).One(&user)
                    
                    if err == nil {
                        
                        // Gravatar url
                        emailHash := gravatar.EmailHash(user.Email)
                        image := gravatar.GetAvatarURL("http", emailHash, "http://spartangeek.com/images/default-avatar.png", 80)
                        
                        // Construct the notification message
                        title := fmt.Sprintf("A **%s** le gusta tu comentario.", user.UserName)
                        message := post.Title
                        
                        // We are inside an isolated routine, so we dont need to worry about the processing cost
                        notify(user_id, "like", post.Id, "/post/" + post.Slug, title, message, image.String())   
                    }
                    
                }(comment_, user_token, post)
                
                c.JSON(200, gin.H{"message": "okay", "status": 606})
                return
            }
        }
    }
    
    c.JSON(401, gin.H{"error": "Couldnt vote, missing information...", "status": 608})
}