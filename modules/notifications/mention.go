package notifications

import (
	"fmt"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/model"
	"gopkg.in/mgo.v2/bson"
	"regexp"
	"strings"
	"time"
)

func (di *NotificationsModule) ParseContentMentions(obj MentionParseObject) {
    
    // Recover from any panic even inside this isolated process 
    defer di.Errors.Recover()
    
    var mention_users, mentions_done []string
    var content string
    var author model.User
    var post model.Post
    
    titles := map[string]string{
        "comment": "**%s** te mencionó en un comentario",
        "post": "**%s** te mencionó en su publicación",
    }
    
    // This is the mention regex to determine the possible users to mention
    mention_regex, _ := regexp.Compile(`(?i)\B\@([\w\-]+)`)
    
    // Services we will need along the runtime
    database := di.Mongo.Database
    firebase := di.Firebase
    
    post = obj.Post
    content = obj.Content
    mention_users = mention_regex.FindAllString(content, -1)
    
    // Get the author of the notification
    err := database.C("users").Find(bson.M{"_id": obj.Author}).One(&author)
    
    if err != nil {
        return
    }
    
    for _, user := range mention_users {
        
        if done, _ := helpers.InArray(user[1:], mentions_done); done == false {
            
            var target_user model.User
            var target_username, target_path, title, message, link string
            var target_notification model.UserFirebaseNotifications
            
            target_username = user[1:]
            
            err := database.C("users").Find(bson.M{"username": target_username}).One(&target_user)
            
            // Don't send the notification if the user has not been found or if the target is the same as the author
            if err != nil || target_user.Id == obj.Author {
                continue
            }
            
            // Replace the mention in the content so it can be a link to the profile
            link = `<a class="user-profile" data-id="`+target_user.Id.Hex()+`">@`+target_username+`</a>`
            content = strings.Replace(content, user, link, -1)
            
            title = fmt.Sprintf(titles[obj.Type], author.UserName)
            message = obj.Title
            
            target_path = "users/" + target_user.Id.Hex() + "/notifications"
            
            // TODO - As the notifications increases this will slow down the whole process, change this
		    target_ref := firebase.Child(target_path, nil, &target_notification)
		    
		    // Increase the notifications count
		    target_ref.Set("count", target_notification.Count+1, nil)
		    
		    // Compose notification
		    notification := &model.UserFirebaseNotification{
    			UserId:       obj.Author,
    			RelatedId:    post.Id,
    			RelatedExtra: post.Slug,
    			Position:     post.Comments.Count,
    			Title:        title,
    			Text:         message,
    			Related:      "mention",
    			Seen:         false,
    			Image:        "",
    			Created:      time.Now(),
    			Updated:      time.Now(),
    		}
    		
    		target_ref.Child("list", nil, nil).Push(notification, nil)
    		
    		// Dont send repeated notifications to the same user even if mentioned twice
    		mentions_done = append(mentions_done, target_username)
        }
    }
    
    if obj.Type == "comment" {
        
        path := "comments.set." + obj.RelatedNested + ".content"
        
        // Compute the database change directives
        change := bson.M{"$set": bson.M{"updated_at": time.Now(), path: content}}   
        
        err = database.C("posts").Update(bson.M{"_id": post.Id}, change)
        
        if err != nil {
            return   
        }
    }
}