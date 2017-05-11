package controller

import (
	"github.com/fernandez14/spartangeek-blacker/core/common"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

// Return all saved macros.
func AllMacros(c *gin.Context) {
	list, err := store.FindMacros(deps.Container, bson.M{})
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, list)
}

// Insert or update macro.
func UpsertMacro(c *gin.Context) {
	var payload store.Macro

	// Validate payload & bind it.
	if err := c.BindJSON(&payload); err != nil {
		c.AbortWithError(400, err)
		return
	}

	macro, err := store.UpsertMacro(deps.Container, payload)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, macro)
}

func DeleteMacro(c *gin.Context) {
	id := bson.ObjectIdHex(c.Param("id"))
	removed, err := store.DeleteMacros(deps.Container, common.ById(id))
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, gin.H{"removed": removed})
}
