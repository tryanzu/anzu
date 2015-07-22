package handle

import (
    "bytes"
    "github.com/fernandez14/spartangeek-blacker/model"
    "github.com/fernandez14/spartangeek-blacker/mongo"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "gopkg.in/mgo.v2/bson"
    "strconv"
    "time"
)

type VoteAPI struct {
    Data        *mongo.Service `inject:""`
	Gaming      *GamingAPI       `inject:""`
}

func (di *VoteAPI) VoteComponent(c *gin.Context) {

    // Get the database interface from the DI
    database := di.Data.Database

    id := c.Params.ByName("id")

    if bson.IsObjectIdHex(id) == false {

        // Invalid request
        c.JSON(400, gin.H{"error": "Invalid request...", "status": 601})

        return
    }

    // Get the user
    user_id := c.MustGet("user_id")
    user_bson_id := bson.ObjectIdHex(user_id.(string))

    // Get the vote content
    var vote model.VoteForm

    if c.BindWith(&vote, binding.JSON) == nil {

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
                    value = 1
                }

                if direction == "down" {
                    value = -1
                }

                // Get the post using the slug
                id := bson.ObjectIdHex(id)
                collection := database.C("posts")

                var post model.Post
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

                    var already_voted model.Vote

                    err = database.C("votes").Find(bson.M{"type": "component", "user_id": user_bson_id, "related_id": id, "nested_type": component}).One(&already_voted)

                    if err == nil {

                        var rem bytes.Buffer

                        // Make the push string
                        rem.WriteString("components.")
                        rem.WriteString(component)
                        rem.WriteString(".votes.")

                        if (direction == "up" && already_voted.Value == 1) || (direction == "down" && already_voted.Value == -1) {

                            rem.WriteString(direction)
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
                    err = collection.Update(bson.M{"_id": post.Id}, change)

                    if err != nil {

                        panic(err)
                    }

                    vote := &model.Vote{
                        UserId:     user_bson_id,
                        Type:       "component",
                        NestedType: component,
                        RelatedId:  id,
                        Value:      value,
                        Created:    time.Now(),
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

func (di *VoteAPI) VoteComment(c *gin.Context) {

    // Get the database interface from the DI
    database := di.Data.Database

    id := c.Params.ByName("id")

    if bson.IsObjectIdHex(id) == false {

        // Invalid request
        c.JSON(400, gin.H{"error": "Invalid request...", "status": 601})

        return
    }

    // Get the user
    user_id := c.MustGet("user_id")
    user_bson_id := bson.ObjectIdHex(user_id.(string))

    // Get the vote content
    var vote model.VoteCommentForm

    if c.BindWith(&vote, binding.JSON) == nil {

        // Get the post using the id
        id := bson.ObjectIdHex(id)
        collection := database.C("posts")

        var post model.Post
        err := collection.FindId(id).One(&post)
        
        if err != nil {
            panic(err)
        }
        
        // Get the author of the vote
        var user model.User
        err = database.C("users").FindId(user_bson_id).One(&user)
        
        if err != nil {
            panic(err)
        }
    
        index := vote.Comment

        if _, err := strconv.Atoi(index); err == nil {
            
            var add bytes.Buffer
            var already_voted model.Vote
            var vote_value int

            err = database.C("votes").Find(bson.M{"type": "comment", "user_id": user_bson_id, "related_id": id, "nested_type": index}).One(&already_voted)
            
            comment_index, _ := strconv.Atoi(index)
            comment_ := post.Comments.Set[comment_index]
            
            if err == nil {
                
                // Cannot allow to change a vote once 15 minutes have passed
                if time.Since(already_voted.Created) > time.Minute * 15 {
                    
                    c.JSON(400, gin.H{"message": "Cannot allow vote changes after 15 minutes.", "status": "error"})
            
                    return
                }
                
                var rem bytes.Buffer

                // Remove the vote from the comment using $inc 
                rem.WriteString("comments.set.")
                rem.WriteString(index)
                
                if already_voted.Value == 1 {
                    
                    rem.WriteString(".votes.up")
                }
                
                if already_voted.Value == -1 {
                  
                    rem.WriteString(".votes.down")
                }
                
                // Comment-To-Change
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
                
                // Return the gamification points
                if already_voted.Value == 1 {
                    
                    go di.Gaming.Related(user_bson_id).RelatedUser(user).Tribute(1)
                    
                    if comment_.Votes.Up - 1 < 5 {
                        
                        go di.Gaming.Related(comment_.UserId).Swords(-1)  
                    }
                    
                    go di.Gaming.Related(comment_.UserId).Coins(-1)  
                    
                } else if already_voted.Value == -1 {
                    
                    go di.Gaming.Related(user_bson_id).RelatedUser(user).Shit(1)
                    
                    if comment_.Votes.Down - 1 < 5 {
                        
                        go di.Gaming.Related(comment_.UserId).Swords(1)  
                    }
                    
                    go di.Gaming.Related(comment_.UserId).Coins(1)  
                }
                
                c.JSON(200, gin.H{"status": "okay"})
                return
            }
            
            // Check if has enough tribute or shit to give    
            if (vote.Direction == "up" && user.Gaming.Tribute < 1) || (vote.Direction == "down" && user.Gaming.Shit < 1) {
                    
                c.JSON(400, gin.H{"message": "Dont have enough gaming points to do this.", "status": "error"})
                
                return
            }
            
            // Make the push string
            add.WriteString("comments.set.")
            add.WriteString(index)
            
            if vote.Direction == "up" {
                
                vote_value = 1
                add.WriteString(".votes.up")
            }
            
            if vote.Direction == "down" {
              
                vote_value = -1
                add.WriteString(".votes.down")
            }

            inc := add.String()
            
            change := bson.M{"$inc": bson.M{inc: 1}}
            err = collection.Update(bson.M{"_id": post.Id}, change)

            if err != nil {
                panic(err)
            }

            vote := &model.Vote{
                UserId:     user_bson_id,
                Type:       "comment",
                NestedType: index,
                RelatedId:  id,
                Value:      vote_value,
                Created:    time.Now(),
            }
            
            err = database.C("votes").Insert(vote)

            // Remove the spend of tribute or shit when giving the vote to the comment (only if comment's user is not the same as the vote's user)
            if comment_.UserId != user_bson_id {
                
                if vote_value == -1 {
                 
                    go di.Gaming.Related(user_bson_id).RelatedUser(user).Shit(-1)  
                    
                    // Remove the swords to the author of the comment
                    if post.Votes.Down <= 5 {
                        
                        go di.Gaming.Related(comment_.UserId).Swords(-1)
                    }
                    
                    go di.Gaming.Related(comment_.UserId).Coins(-1)
                    
                } else {
                    
                    go di.Gaming.Related(user_bson_id).RelatedUser(user).Tribute(-1)     
                    
                    // Give the swords to the author of the comment
                    if post.Votes.Up <= 5 {
                 
                        go di.Gaming.Related(comment_.UserId).Swords(1)
                    }
                    
                    go di.Gaming.Related(comment_.UserId).Coins(1)
                }
            }
            
            // Notify the author of the comment
            go func(comment model.Comment, token bson.ObjectId, post model.Post) {

                /*user_id := comment.UserId

                  // Get the comment like author
                  var user model.User

                  database.C("users").Find(bson.M{"_id": token.UserId}).One(&user)

                  if err == nil {

                      // Gravatar url
                      emailHash := gravatar.EmailHash(user.Email)
                      image := gravatar.GetAvatarURL("http", emailHash, "http://spartangeek.com/images/default-avatar.png", 80)

                      // Construct the notification message
                      title := fmt.Sprintf("A **%s** le gusta tu comentario.", user.UserName)
                      message := post.Title

                      // We are inside an isolated routine, so we dont need to worry about the processing cost
                      //notify(user_id, "like", post.Id, "/post/" + post.Slug, title, message, image.String())
                  }*/

            }(comment_, user_bson_id, post)

            c.JSON(200, gin.H{"status": "okay"})
            return
        }
    }

    c.JSON(401, gin.H{"error": "Couldnt vote, missing information...", "status": 608})
}