package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/flags"
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

type upsertFlagForm struct {
	RelatedTo string        `json:"related_to" binding:"required,eq=post|eq=comment"`
	RelatedID bson.ObjectId `json:"related_id" binding:"required"`
	Reason    string        `json:"category" binding:"required"`
	Content   string        `json:"content" binding:"max=255"`
}

// NewFlag endpoint.
func NewFlag(c *gin.Context) {
	var form upsertFlagForm
	if err := c.BindJSON(&form); err != nil {
		jsonBindErr(c, http.StatusBadRequest, "Invalid flag request, check parameters", err)
		return
	}

	rules := config.C.Rules()
	if _, exists := rules.Flags[form.Reason]; false == exists {
		jsonErr(c, http.StatusBadRequest, "Invalid flag reason")
		return
	}

	usr := c.MustGet("user").(user.User)
	if count := flags.TodaysCountByUser(deps.Container, usr.Id); count > 10 {
		jsonErr(c, http.StatusPreconditionFailed, "Can't flag anymore for today")
		return
	}

	flag, err := flags.UpsertFlag(deps.Container, flags.Flag{
		UserID:    usr.Id,
		RelatedID: &form.RelatedID,
		RelatedTo: form.RelatedTo,
		Content:   form.Content,
		Reason:    form.Reason,
	})
	if err != nil {
		jsonErr(c, http.StatusInternalServerError, err.Error())
		return
	}

	events.In <- events.NewFlag(flag.ID)
	c.JSON(200, gin.H{"flag": flag, "status": "okay"})
}

// Flag status request.
func Flag(c *gin.Context) {
	var (
		id      bson.ObjectId
		related = c.Params.ByName("related")
	)
	if id = bson.ObjectIdHex(c.Params.ByName("id")); !id.Valid() {
		jsonErr(c, http.StatusBadRequest, "malformed request, invalid id")
		return
	}
	usr := c.MustGet("user").(user.User)
	f, err := flags.FindOne(deps.Container, related, id, usr.Id)
	if err != nil {
		jsonErr(c, http.StatusNotFound, "flag not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"flag": f})
}

// FlagReasons endpoint.
func FlagReasons(c *gin.Context) {
	rules := config.C.Rules()
	reasons := []string{}
	for k := range rules.Flags {
		reasons = append(reasons, k)
	}
	c.JSON(200, gin.H{"status": "okay", "reasons": reasons})
}
