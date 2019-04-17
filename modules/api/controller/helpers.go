package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/modules/acl"
	"gopkg.in/go-playground/validator.v8"
)

func jsonErr(c *gin.Context, status int, message string) {
	// This specific json error structure is handled
	// by the frontend in a generic way so errors
	// can be shown to the user and also translated.
	c.AbortWithStatusJSON(status, gin.H{
		"status":  "error",
		"message": message,
	})
}

func jsonBindErr(c *gin.Context, status int, message string, bindErr error) {
	// This specific json error structure is handled
	// by the frontend in a generic way so errors
	// can be shown to the user and also translated.
	c.AbortWithStatusJSON(status, gin.H{
		"status":  "error",
		"message": message,
		"details": bindErr.(validator.ValidationErrors),
	})
}

func perms(c *gin.Context) *acl.User {
	return c.MustGet("acl").(*acl.User)
}
