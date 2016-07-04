package comments

import (
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/transmit"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Add(c *gin.Context) {

	var form CommentForm

	id := c.Params.ByName("id")

	if bson.IsObjectIdHex(id) == false {
		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	post, err := this.Feed.Post(bson.ObjectIdHex(id))

	if err != nil {
		c.JSON(404, gin.H{"status": "error", "message": "Post not found."})
		return
	}

	user_str := c.MustGet("user_id")
	user_id := bson.ObjectIdHex(user_str.(string))

	if c.Bind(&form) == nil {

		if post.IsLocked() {
			c.JSON(403, gin.H{"status": "error", "message": "Comments not longer allowed in this post."})
			return
		}

		comment := post.PushComment(form.Content, user_id)

		go func(carrier *transmit.Sender, id bson.ObjectId, usrId bson.ObjectId) {

			carrierParams := map[string]interface{}{
				"fire":    "new-comment",
				"id":      id.Hex(),
				"user_id": usrId.Hex(),
			}

			carrier.Emit("feed", "action", carrierParams)

		}(this.Transmit, post.Id, user_id)

		if post.UserId != user_id {

			go func(post *feed.Post, comment *feed.Comment, user_id bson.ObjectId) {

				if post.Category.Hex() != "55dc16593f6ba1005d000007" {
					// Tell the new comment for gamification
					this.Gaming.Get(user_id).Did("comment")
				}

				// Notify the author about this comment
				this.Notifications.Comment(post.Slug, post.Title, comment.Position, post.Id, post.UserId, user_id)

			}(post, comment, user_id)
		}

		/*// Update the post and push the comments
		change := bson.M{"$push": bson.M{"comments.set": comment}, "$set": bson.M{"updated_at": time.Now()}, "$inc": bson.M{"comments.count": 1}}
		err = database.C("posts").Update(bson.M{"_id": post.Id}, change)

		if err != nil {
			panic(err)
		}

		position := strconv.Itoa(len(post.Comments.Set))

		// Process the mentions. TODO - Determine race conditions
		go di.Notifications.ParseContentMentions(notifications.MentionParseObject{
			Type:          "comment",
			RelatedNested: position,
			Content:       comment.Content,
			Title:         post.Title,
			Author:        user_bson_id,
			Post:          post,
		})

		go func(carrier *transmit.Sender, id bson.ObjectId, usrId bson.ObjectId) {

			carrierParams := map[string]interface{}{
				"fire":    "new-comment",
				"id":      id.Hex(),
				"user_id": usrId.Hex(),
			}

			carrier.Emit("feed", "action", carrierParams)

		}(di.Transmit, post.Id, user_bson_id)

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
			err = database.C("posts").Update(bson.M{"_id": post.Id}, change)

			if err != nil {
				panic(err)
			}
		}

		// Triggers when the author of the comment is not the post's author
		if post.UserId != user_bson_id {

			// Notify the post's author
			go di.notifyCommentPostAuth(post, user_bson_id)

			// Add the gamification contribution
			go di.Gaming.Get(user_bson_id).Did("comment")
		}*/

		c.JSON(200, gin.H{"status": "okay", "message": comment.Content, "position": comment.Position})
		return
	}

	c.JSON(401, gin.H{"error": "Not authorized.", "status": 704})
}
