package handle

import (
    "github.com/fernandez14/spartangeek-blacker/mongo"
    "github.com/fernandez14/spartangeek-blacker/model"
    "github.com/cosn/firebase"
    "gopkg.in/mgo.v2/bson"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "github.com/ftrvxmtrx/gravatar"
    "time"
    "regexp"
    "strings"
    "fmt"
)

type CommentAPI struct {
    DataService  *mongo.Service `inject:""`
    Firebase *firebase.Client `inject:""`
}

func (di *CommentAPI) CommentAdd (c *gin.Context) {

    // Get the database interface from the DI
    database := di.DataService.Database

    id := c.Params.ByName("id")

    if bson.IsObjectIdHex(id) == false {

        c.JSON(400, gin.H{"error": "Invalid request, no valid params.", "status": 701})
        return
    }

    user_id := c.MustGet("user_id")
    user_bson_id := bson.ObjectIdHex(user_id.(string))

    var comment model.CommentForm

    if c.BindWith(&comment, binding.JSON) == nil {

        // Get the post using the slug
        id := bson.ObjectIdHex(id)
    
        // Posts collection
        collection := database.C("posts")
        
        var post model.Post
        
        err := collection.FindId(id).One(&post)
        
        if err != nil {
            
            c.JSON(404, gin.H{"error": "Couldnt find the post", "status": 705})
            return
        }
        
        votes := model.Votes{
            Up: 0,
            Down: 0,
        }
        
        comment := model.Comment{
            UserId: user_bson_id,
            Votes: votes,
            Content: comment.Content,
            Created: time.Now(),
        }
        
        //urls, _  := regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
        mentions, _ := regexp.Compile(`(?i)\@[a-z0-9\-\_]+`)
            
        //var assets []string
        
        //assets = urls.FindAllString(comment.Content, -1)
        
        //for _, asset := range assets {

            // Download the asset on other routine in order to non block the API request
            //go downloadFromUrl(asset)
        //}
        
        var mentions_users []string
        
        mentions_users = mentions.FindAllString(comment.Content, -1)
        
        for _, mention_user := range mentions_users {
            
            go di.mentionUserComment(mention_user, post, user_bson_id)    
            
            // Replace the mentioned user
            markdown := "[" + mention_user + "](/user/profile/" + mention_user[1:] + " \""+ mention_user[1:] +"\")"  
            comment.Content =  strings.Replace(comment.Content, mention_user, markdown, -1)
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

            if already_within == user_bson_id {

                need_add = false
            }
        }

        if need_add == true {

            // Add the user to the user list
            change := bson.M{"$push": bson.M{"users": user_bson_id}}
            err = collection.Update(bson.M{"_id": post.Id}, change)
            
            if err != nil {
    		    panic(err)
    	    }
        }
        
        // Notifications for the author 
        if post.UserId != user_bson_id {

            go di.notifyCommentPostAuth(post, post.UserId)
        }

        c.JSON(200, gin.H{"message": "okay", "status": 706})
        return
    }
    
    c.JSON(401, gin.H{"error": "Not authorized.", "status": 704})
}

func (di *CommentAPI) notifyCommentPostAuth(post model.Post, user_id bson.ObjectId) {

    // Get the comment author
    var user model.User
    var notifications model.UserFirebaseNotifications
    
    err := di.DataService.Database.C("users").Find(bson.M{"_id": user_id}).One(&user)
    
    if err == nil {
        
        // Gravatar url
        emailHash := gravatar.EmailHash(user.Email)
        image := gravatar.GetAvatarURL("http", emailHash, "http://spartangeek.com/images/default-avatar.png", 80)
        
        // Construct the notification message
        title := fmt.Sprintf("Nuevo comentario de **%s**", user.UserName)
        message := post.Title
        
        // Process notification using firebase
        authorPath := "users/" + user_id.Hex() + "/notifications"
        authorRef  := di.Firebase.Child(authorPath, nil, &notifications)

        authorRef.Set("count", notifications.Count + 1, nil)

        notification := &model.UserFirebaseNotification {
            UserId: post.UserId,
            RelatedId: post.Id,
            Title: title,
            Text: message,
            Related: "comment",
            Seen: false,
            Image: image.String(),
            Created: time.Now(),
            Updated: time.Now(),
        }

        authorRef.Child("list", nil, nil).Push(notification, nil)
    }       
}

func (di *CommentAPI) mentionUserComment(mentioned string, post model.Post, user_id bson.ObjectId) {
    
    /*var user User
    var author User
    
    username := mentioned[1:]
    
    err := database.C("users").Find(bson.M{"username": username}).One(&user)
    
    if err == nil {
        
        err := database.C("users").Find(bson.M{"_id": user_id}).One(&author)
        
        if err == nil {
            
            // Gravatar url
            emailHash := gravatar.EmailHash(author.Email)
            image := gravatar.GetAvatarURL("http", emailHash, "http://spartangeek.com/images/default-avatar.png", 80)
            
            // Construct the notification message
            title := fmt.Sprintf("**%s** te menciono en un comentario", author.UserName)
            message := post.Title
            
            // We are inside an isolated routine, so we dont need to worry about the processing cost
            notify(user.Id, "mention", post.Id, "/post/" + post.Slug, title, message, image.String())
        }
    }*/
}