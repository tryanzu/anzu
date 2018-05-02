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
	if usr.Gaming.Swords < 15 {
		c.JSON(http.StatusPreconditionFailed, gin.H{"message": "Not enough user reputation.", "status": "error"})
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
}
