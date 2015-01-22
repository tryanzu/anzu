package main

import (
    "gopkg.in/mgo.v2/bson"
    "github.com/gin-gonic/gin"
    "encoding/json"
    "time"
    "strings"
    "strconv"
)

type Counter struct {
    UserId   bson.ObjectId `bson:"user_id" json:"user_id"`
    Counters map[string]PostCounter `bson:"counters" json:"counters"`
}

type PostCounter struct {
    Counter int `bson:"counter" json:"counter"`
    Updated time.Time `bson:"updated_at" json:"updated_at"`
}

type Notification struct {
    Id        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
    UserId    bson.ObjectId `bson:"user_id" json:"user_id"`
    RelatedId bson.ObjectId `bson:"related_id" json:"related_id"`
    Title     string `bson:"title" json:"title"`
    Text      string `bson:"text" json:"text"`
    Link      string `bson:"link" json:"link"`
    Related   string `bson:"related" json:"related"`
    Seen      bool   `bson:"seen" json:"seen"`
    Image     string `bson:"image" json:"image"`
    Created   time.Time `bson:"created_at" json:"created_at"`
    Updated   time.Time `bson:"updated_at" json:"updated_at"`
}

func notify(to bson.ObjectId, of string, related bson.ObjectId, link string, title string, text string, image string) {
    
    notification := &Notification {
        UserId: to,
        RelatedId: related,
        Title: title,
        Text: text,
        Link: link,
        Related: of,
        Seen: false,
        Image: image,
        Created: time.Now(),
        Updated: time.Now(),
    }
    
    err := database.C("notifications").Insert(notification)
    
    if err != nil {
        panic(err)
    }
    
    // Send the socket notification (real-time)
    realtime := map[string] string {
        "perform": "notification",
        "to": "user.bulletin:" + to.Hex(),
    }
    
    realtime_message, _ := json.Marshal(realtime)
    zmq_messages <- string(realtime_message)
    
    return
}

func UserNotificationsGet(c *gin.Context) {
    
    // Get user by token
	user_token := UserToken{}
	token := c.Request.Header.Get("Auth-Token")
	qs := c.Request.URL.Query()
	
	if token != "" {
	    
		err  := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)
		
		if err == nil {
		    
		    offset := 0
	        limit  := 10
	
	        o := qs.Get("offset")
	        l := qs.Get("limit")
	        
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
	
		    var notifications []Notification
		    
		    // Get the user notifications
			err := database.C("notifications").Find(bson.M{"user_id": user_token.UserId}).Limit(limit).Skip(offset).All(&notifications)
			
		    if err != nil {
		        panic(err)
		    }
		    
		    // Reset the user notifications counter while sending the notifications
		    go func(user_id bson.ObjectId) {
		        
		        // Update the collection of counters
                database.C("notifications").UpdateAll(bson.M{"user_id": user_id}, bson.M{"$set": bson.M{"seen": true, "updated_at": time.Now()}})
                
                // Send the socket notification (real-time)
                realtime := map[string] string {
                    "perform": "notification-clear",
                    "to": "user.bulletin:" + user_id.Hex(),
                }
                
                realtime_message, _ := json.Marshal(realtime)
                zmq_messages <- string(realtime_message)
    
		    }(user_token.UserId)
		    
		    if len(notifications) > 0 {
		        
		        c.JSON(200, gin.H{"notifications": notifications, "offset": offset, "limit": limit})
		    } else {
		        
		        c.JSON(200, gin.H{"notifications": []string{}, "offset": offset, "limit": limit})
		    }
		    return
		}
	}
	
	c.JSON(401, gin.H{"message": "No auth credentials", "status": "error", "code": 401})
}

func counterAdd(category string) {
    
    // Replace the slug dash with underscore 
    counter := strings.Replace(category, "-", "_", -1)
    find := "counters." + counter + ".counter"
    
    // Update the collection of counters
    database.C("counters").UpdateAll(nil, bson.M{"$inc": bson.M{find: 1}})
    
    return
}

func counterReset(category string, user_id bson.ObjectId) {
    
    // Replace the slug dash with underscore 
    counter := strings.Replace(category, "-", "_", -1)
    find := "counters." + counter + ".counter"
    updated_at := "counters." + counter + ".updated_at"
    
    // Update the collection of counters
    err := database.C("counters").Update(bson.M{"user_id": user_id}, bson.M{"$set": bson.M{find: 0, updated_at: time.Now()}})
    
    if err != nil {
        panic(err)   
    }
    
    return
}