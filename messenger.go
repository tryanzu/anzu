package main

import (
    "gopkg.in/mgo.v2/bson"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "time"
    "regexp"
    "strconv"
)


type Message struct {
    Id         bson.ObjectId   `bson:"_id,omitempty" json:"id,omitempty"`
    UserId     bson.ObjectId   `bson:"user_id" json:"user_id"`
    Message    string          `bson:"message" json:"message"`
    Author     User            `bson:"author,omitempty" json:"author,omitempty"`
    Created    time.Time       `bson:"created_at" json:"created_at"`
}

type Hashtag struct {
	Id      	bson.ObjectId   	`bson:"_id,omitempty" json:"id,omitempty"`
	Name    	string 				`bson:"name" json:"name"`
	UserId  	bson.ObjectId 		`bson:"user_id" json:"user_id"`
	Mentions  	int 				`bson:"mentions" json:"mentions"`
	List        []HashtagMention 	`bson:"list" json:"list"`
	Created 	time.Time 			`bson:"created_at" json:"created_at"`
	Updated 	time.Time 			`bson:"updated_at" json:"updated_at"`
}

type HashtagMention struct {
	UserId 		bson.ObjectId 	`bson:"user_id" json:"user_id"`
	Created     time.Time 		`bson:"created_at" json:"created_at"`
}

type MessageForm struct {
    Message string `json:"message" binding:"required"`    
}

func MessagesGet(c *gin.Context) {
    
    var list []Message
	offset := 0
	limit := 30
	query := bson.M{}

	qs := c.Request.URL.Query()

	o := qs.Get("offset")
	l := qs.Get("limit")
	h := qs.Get("lookup")
	
	// Check if offset has been specified
	if o != "" {
		off, err := strconv.Atoi(o)

		if err != nil || off < 0 {
			c.JSON(401, gin.H{"message": "Invalid request, check params.", "status": "error", "code": 901})
			return
		}

		offset = off
	}

	// Check if limit has been specified
	if l != "" {
		lim, err := strconv.Atoi(l)

		if err != nil || lim <= 0 {
			c.JSON(401, gin.H{"message": "Invalid request, check params.", "status": "error", "code": 901})
			return
		}

		limit = lim
	}
	
	if h != "" {
		
		// Lookup for the hashtag using an full-text search index
		query = bson.M{"$text": bson.M{"$search": h}}
	}
	
	err := database.C("chat").Find(query).Sort("-created_at").Limit(limit).Skip(offset).All(&list)
	
	if err != nil {
	    panic(err)    
	}
	
	var authors []bson.ObjectId
	var users []User
	
	for _, message := range list {
	    
	    authors = append(authors, message.UserId)
	}
	
	// Get the authors of the messages
	err = database.C("users").Find(bson.M{"_id": bson.M{"$in": authors}}).All(&users)
    
    if err != nil {
        panic(err)
    }
    
    if len(list) > 0 {
        
        usersMap := make(map[bson.ObjectId]User)

		for _, user := range users {

			usersMap[user.Id] = user
		}
		
		for index := range list {

			msg := &list[index]

			if _, okay := usersMap[msg.UserId]; okay {

				msgUser := usersMap[msg.UserId]

				msg.Author = User{
					Id:        msgUser.Id,
					UserName:  msgUser.UserName,
					FirstName: msgUser.FirstName,
					LastName:  msgUser.LastName,
					Email:     msgUser.Email,
				}
			}
		}
		
		c.JSON(200, gin.H{"messages": list, "offset": offset, "limit": limit})
    } else {
        
        c.JSON(200, gin.H{"messages": []string{}, "offset": offset, "limit": limit})   
    }
}

func MessagePublish(c *gin.Context) {
    
    // Get user by token
	user_token := UserToken{}
	token := c.Request.Header.Get("Auth-Token")
	
	if token == "" {
	    
	    c.JSON(401, gin.H{"message": "No auth credentials", "status": "error", "code": 401})
		return   
	}
	
	// Try to fetch the user using token header though
	err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)

	if err != nil {

		c.JSON(401, gin.H{"message": "No valid auth credentials", "status": "error", "code": 401})
		return
	}
	
	var input MessageForm
	
	if c.BindWith(&input, binding.JSON) {
	    
		// Look up for hashtags inside the published message
		hashtag_regex, _ := regexp.Compile(`(?i)\#[0-9a-z\_\-]+`)
		hashtag := hashtag_regex.FindAllString(input.Message, -1)
		
	    message := &Message{
	        UserId:  user_token.UserId,
	        Message: input.Message,
	        Created: time.Now(),
	    }
		
	    // Spread through the socket
	    go func(token UserToken, message string) {
	        
	        // Get the username
	        var user User
	        
	        err := database.C("users").Find(bson.M{"_id": token.UserId}).One(&user)
	        
	        if err == nil {
	            
	            // Spread the message to the main chat 
	            send := map[string]string {
	                "message": message,
	                "username": user.UserName,
	                "user_id": user.Id.Hex(),
	            }
	            
	            spread("messaging:main", send)   
	        }
	        
	    }(user_token, input.Message)
		
		err = database.C("chat").Insert(message)
		
		if err != nil {
			panic(err)
		}
		
		for _, hashtag_matched := range hashtag {
			
			// Do the hashtags calculations
			go func(hashtag string, token UserToken) {
				
				var hashtag_lookup Hashtag
				
				// Check if hashtag already in collection
				err := database.C("hashtags").Find(bson.M{"name": hashtag[1:]}).One(&hashtag_lookup)
				
				// Create or update the list of mentions
				if err != nil {
					
					new_hashtag := &Hashtag{
						Name: hashtag[1:],
						UserId: token.UserId,
						Mentions: 1,
						List: make([]HashtagMention, 0),
						Created: time.Now(),
						Updated: time.Now(),
					}
					
					// Insert onto the list of hashtags no mattering if succedd or not
					database.C("hashtags").Insert(new_hashtag)
					
				} else {
					
					// Push a mention and increase the number of mentions
					mention := &HashtagMention{
						UserId: token.UserId,
						Created: time.Now(),
					}
					
					change := bson.M{"$push": bson.M{"list": mention}, "$inc": bson.M{"mentions": 1}, "$set": bson.M{"updated_at": time.Now()}}
					
					// Update the hashtag no panicing if failed
					database.C("hashtags").Update(bson.M{"_id": hashtag_lookup.Id}, change)
				}
				
			}(hashtag_matched, user_token)
		}
		
		c.JSON(200, gin.H{"status": "okay", "code": 200})
	}
}

func HashtagsGet(c *gin.Context) {
	
	var hashtags []Hashtag
	
	// Only select the needed fields, the list of mentions is not required(may be too expensive)
	fields := bson.M{"_id": 1, "name": 1, "mentions": 1, "created_at": 1, "updated_at": 1, "user_id": 1}
	
	// Get the hashtags sorting by the newests with more mentions
	err := database.C("hashtags").Find(bson.M{}).Select(fields).Sort("-mentions, -updated_at").Limit(10).All(&hashtags)
	
	if err == nil {
		
		// Print out the list of hashtags
		c.JSON(200, gin.H{"hashtags": hashtags})
	} else {
		
		// Print out an empty list of hashtags
		c.JSON(200, gin.H{"hashtags": []string{}})
	}
}