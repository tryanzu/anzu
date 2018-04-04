package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/core/config"
)

// UpdateConfig from running anzu.
func UpdateConfig(c *gin.Context) {
	var update ConfigUpdate
	if err := c.Bind(&update); err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("Invalid payload."))
		return
	}

	//usr := c.MustGet("user").(user.User)
	err := config.MergeUpdate(map[string]interface{}{
		update.Section: update.Changes,
	})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "okay"})
}

type ConfigUpdate struct {
	Section string                 `json:"section" binding:"required"`
	Changes map[string]interface{} `json:"changes" binding:"required"`
}
