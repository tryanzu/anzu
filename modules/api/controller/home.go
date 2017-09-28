package controller

import (
	"github.com/gin-gonic/gin"
)

func HomePage(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.HTML(200, "pages/home.tmpl", gin.H{
		"title": "Main website",
	})
}
