package main

import (
    "gopkg.in/mgo.v2/bson"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "time"
    "regexp"
)

type CommentForm struct {
    Content   string `json:"content" binding:"required"`
}

func CommentAdd (c *gin.Context) {

    id := c.Params.ByName("id")

    if bson.IsObjectIdHex(id) == false {

        c.JSON(400, gin.H{"error": "Invalid request, no valid params.", "status": 701})
        return
    }

    // Get the query parameters
    qs := c.Request.URL.Query()

    // Name of the set to get
    token := qs.Get("token")

    if token == "" {

        c.JSON(401, gin.H{"error": "Not authorized.", "status": 702})
        return
    }

    // Get user by token
    user_token  := UserToken{}

    // Try to fetch the user using token header though
	err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)

	if err != nil {
	    c.JSON(401, gin.H{"error": "Not authorized.", "status": 703})
        return
	}

    var comment CommentForm

    if c.BindWith(&comment, binding.JSON) {

        // Get the post using the slug
        id := bson.ObjectIdHex(id)
    
        // Posts collection
        collection := database.C("posts")
        
        var post Post
        
        err := collection.FindId(id).One(&post)
        
        if err != nil {
            
            c.JSON(404, gin.H{"error": "Couldnt find the post", "status": 705})
            return
        }
        
        votes := Votes{
            Up: 0,
            Down: 0,
        }
        
        comment := Comment{
            UserId: user_token.UserId,
            Votes: votes,
            Content: comment.Content,
            Created: time.Now(),
        }
        
        urls, _ := regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
        
        var assets []string
        
        assets = urls.FindAllString(comment.Content, -1)
        
        for _, asset := range assets {

            // Download the asset on other routine in order to non block the API request
            go downloadFromUrl(asset)
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

        c.JSON(200, gin.H{"message": "okay", "status": 706})
        return
    }
    
    c.JSON(401, gin.H{"error": "Not authorized.", "status": 704})
}