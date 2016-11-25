package votes

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/transmit"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

	"strconv"
	"time"
)

func (this API) Comment(c *gin.Context) {

	database := this.Mongo.Database
	id := c.Params.ByName("id")

	if bson.IsObjectIdHex(id) == false {
		c.JSON(400, gin.H{"message": "Malformed request, invalid id.", "status": "error"})
		return
	}

	user_str := c.MustGet("user_id")
	user_id := bson.ObjectIdHex(user_str.(string))

	var vote CommentForm

	if c.Bind(&vote) == nil {

		id := bson.ObjectIdHex(id)
		comment, err := this.Feed.GetComment(id)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": err.Error()})
			return
		}

		post := comment.GetPost()

		// Get the author of the vote
		usr, err := this.User.Get(user_id)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": err.Error()})
			return
		}

		usr_model := usr.Data()

		var alreadyVoted model.Vote
		var voteValue int

		err = database.C("votes").Find(bson.M{"type": "comment", "user_id": user_id, "related_id": post.Id, "nested_type": strconv.Itoa(comment.Position)}).One(&alreadyVoted)

		if err == nil {

			// Cannot allow to change a vote once 15 minutes have passed
			if time.Since(alreadyVoted.Created) > time.Minute*15 {
				c.JSON(400, gin.H{"message": "Cannot allow vote changes after 15 minutes.", "status": "error"})
				return
			}

			var mutator string

			if alreadyVoted.Value == 1 {
				mutator = "votes.up"
			} else {
				mutator = "votes.down"
			}

			err := database.C("comments").Update(bson.M{"_id": comment.Id}, bson.M{"$inc": bson.M{mutator: -1}})

			if err != nil {
				panic(err)
			}

			err = database.C("votes").RemoveId(alreadyVoted.Id)

			if err != nil {
				c.JSON(409, gin.H{"message": "Could not found vote reference.", "status": "error"})
				return
			}

			// Return the gamification points
			if alreadyVoted.Value == 1 {

				go func(usr *user.One, comment_owner bson.ObjectId) {

					this.Gaming.Get(usr).Tribute(1)

					author := this.Gaming.Get(comment_owner)
					author.Coins(-1)

					if comment.Votes.Up-1 < 5 {
						author.Swords(-1)
					}

				}(usr, comment.UserId)

				go func(carrier *transmit.Sender, id, cid bson.ObjectId, pos int) {

					carrierParams := map[string]interface{}{
						"fire":  "comment-upvote-remove",
						"id":    cid.Hex(),
						"index": pos,
					}

					carrier.Emit("post", id.Hex(), carrierParams)

				}(this.Transmit, post.Id, comment.Id, comment.Position)

			} else if alreadyVoted.Value == -1 {

				go func(usr *user.One, comment_owner bson.ObjectId) {

					this.Gaming.Get(usr).Shit(1)

					author := this.Gaming.Get(comment_owner)
					author.Coins(1)

					if comment.Votes.Down-1 < 5 {
						author.Swords(1)
					}

				}(usr, comment.UserId)

				go func(carrier *transmit.Sender, id, cid bson.ObjectId, pos int) {

					carrierParams := map[string]interface{}{
						"fire":  "comment-downvote-remove",
						"id":    cid.Hex(),
						"index": pos,
					}

					carrier.Emit("post", id.Hex(), carrierParams)

				}(this.Transmit, post.Id, comment.Id, comment.Position)
			}

			c.JSON(200, gin.H{"status": "okay"})
			return
		}

		// Check if has enough tribute or shit to give
		if (vote.Direction == "up" && usr_model.Gaming.Tribute < 1) || (vote.Direction == "down" && usr_model.Gaming.Shit < 1) {
			c.JSON(400, gin.H{"message": "Dont have enough gaming points to do this.", "status": "error"})
			return
		}

		var mutator string

		if vote.Direction == "up" {
			voteValue = 1
			mutator = "votes.up"

			go func(carrier *transmit.Sender, id, cid bson.ObjectId, pos int) {

				carrierParams := map[string]interface{}{
					"fire":  "comment-upvote",
					"id":    cid.Hex(),
					"index": pos,
				}

				carrier.Emit("post", id.Hex(), carrierParams)

			}(this.Transmit, post.Id, comment.Id, comment.Position)
		}

		if vote.Direction == "down" {

			voteValue = -1
			mutator = "votes.down"

			go func(carrier *transmit.Sender, id, cid bson.ObjectId, pos int) {

				carrierParams := map[string]interface{}{
					"fire":  "comment-downvote",
					"id":    cid.Hex(),
					"index": pos,
				}

				carrier.Emit("post", id.Hex(), carrierParams)

			}(this.Transmit, post.Id, comment.Id, comment.Position)
		}

		err = database.C("comments").Update(bson.M{"_id": comment.Id}, bson.M{"$inc": bson.M{mutator: 1}})

		if err != nil {
			panic(err)
		}

		vote := &model.Vote{
			UserId:     user_id,
			Type:       "comment",
			NestedType: strconv.Itoa(comment.Position),
			RelatedId:  post.Id,
			Value:      voteValue,
			Created:    time.Now(),
		}

		err = database.C("votes").Insert(vote)

		// Remove the spend of tribute or shit when giving the vote to the comment (only if comment's user is not the same as the vote's user)
		if comment.UserId != user_id {

			if voteValue == -1 {

				go func(usr *user.One, comment *feed.Comment) {

					this.Gaming.Get(usr).Shit(-1)

					author := this.Gaming.Get(comment.UserId)
					author.Coins(-1)

					if comment.Votes.Down <= 5 {
						author.Swords(-1)
					}

				}(usr, comment)

			} else {

				go func(usr *user.One, comment *feed.Comment) {

					this.Gaming.Get(usr).Tribute(-1)

					author := this.Gaming.Get(comment.UserId)
					author.Coins(1)

					if comment.Votes.Up <= 5 {
						author.Swords(1)
					}

				}(usr, comment)
			}
		}

		c.JSON(200, gin.H{"status": "okay"})
		return
	}

	c.JSON(401, gin.H{"error": "Couldnt vote, missing information...", "status": 608})
}
