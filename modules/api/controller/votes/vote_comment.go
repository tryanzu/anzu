package votes

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/comments"
	"github.com/tryanzu/core/board/votes"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"

	"net/http"
)

// Comment vote delivery.
func (api API) Comment(c *gin.Context) {
	var (
		id      bson.ObjectId
		form    CommentForm
		comment comments.Comment
		err     error
	)

	usr := c.MustGet("user").(user.User)
	if usr.Gaming.Tribute < 1 {
		c.JSON(http.StatusPreconditionFailed, gin.H{"message": "Not enough user vote points.", "status": "error"})
		return
	}

	// Comment id validation.
	if id = bson.ObjectIdHex(c.Params.ByName("id")); !id.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Malformed request, invalid id.", "status": "error"})
		return
	}

	// Bind form data.
	if err = c.Bind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "reason": "Invalid request."})
		return
	}

	if comment, err = comments.FindId(deps.Container, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "reason": "Invalid id."})
		return
	}

	vote, err := votes.UpsertVote(deps.Container, comment, usr.Id, form.VoteType())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Events pool signal
	events.In <- events.VoteComment(vote)

	if vote.Deleted != nil {
		c.JSON(http.StatusOK, gin.H{"action": "delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"action": "create"})

	/*
		comment, err := this.Feed.GetComment(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
			return
		}

		post := comment.GetPost()

		// Get the author of the vote
		usr, err := this.User.Get(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
			return
		}

		usrModel := usr.Data()

		var (
			alreadyVoted model.Vote
			voteValue int
		)

		err = database.C("votes").Find(bson.M{
			"type":        "comment",
			"user_id":     userID,
			"related_id":  post.Id,
			"nested_type": strconv.Itoa(comment.Position),
		}).One(&alreadyVoted)

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

				events.In <- events.RawEmit("post", post.Id.Hex(), map[string]interface{}{
					"fire":  "comment-upvote-remove",
					"id":    comment.Id.Hex(),
					"index": comment.Position,
				})

			} else if alreadyVoted.Value == -1 {

				go func(usr *user.One, comment_owner bson.ObjectId) {

					this.Gaming.Get(usr).Shit(1)

					author := this.Gaming.Get(comment_owner)
					author.Coins(1)

					if comment.Votes.Down-1 < 5 {
						author.Swords(1)
					}

				}(usr, comment.UserId)

				events.In <- events.RawEmit("post", post.Id.Hex(), map[string]interface{}{
					"fire":  "comment-downvote-remove",
					"id":    comment.Id.Hex(),
					"index": comment.Position,
				})
			}

			c.JSON(200, gin.H{"status": "okay"})
			return
		}

		// Check if has enough tribute or shit to give
		if (form.Direction == "up" && usrModel.Gaming.Tribute < 1) || (form.Direction == "down" && usrModel.Gaming.Shit < 1) {
			c.JSON(400, gin.H{"message": "Dont have enough gaming points to do this.", "status": "error"})
			return
		}

		var mutator string

		if form.Direction == "up" {
			voteValue = 1
			mutator = "votes.up"

			events.In <- events.RawEmit("post", post.Id.Hex(), map[string]interface{}{
				"fire":  "comment-upvote",
				"id":    comment.Id.Hex(),
				"index": comment.Position,
			})
		}

		if form.Direction == "down" {
			voteValue = -1
			mutator = "votes.down"
			events.In <- events.RawEmit("post", post.Id.Hex(), map[string]interface{}{
				"fire":  "comment-downvote",
				"id":    comment.Id.Hex(),
				"index": comment.Position,
			})
		}

		err = database.C("comments").Update(bson.M{"_id": comment.Id}, bson.M{"$inc": bson.M{mutator: 1}})

		if err != nil {
			panic(err)
		}

		vote := &model.Vote{
			UserId:     userID,
			Type:       "comment",
			NestedType: strconv.Itoa(comment.Position),
			RelatedId:  post.Id,
			Value:      voteValue,
			Created:    time.Now(),
		}

		err = database.C("votes").Insert(vote)

		// Remove the spend of tribute or shit when giving the vote to the comment (only if comment's user is not the same as the vote's user)
		if comment.UserId != userID {
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

		c.JSON(200, gin.H{"status": "okay"})*/
}
