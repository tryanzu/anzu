package controller

import (
	"github.com/gin-gonic/gin"
	"strings"
)

type GeneralAPI struct {
	Page string
}

func (this GeneralAPI) Landing(c *gin.Context) {
	
	var page string = this.Page

	page = strings.Replace(page, "SpartanGeek.com | Comunidad de tecnología, geeks y más", "SpartanGeek.com | Comunidad de Videojuegos, PC, Tecnología y más", 1)
	page = strings.Replace(page, "{{ page.title }}", "SpartanGeek | Comunidad de Videojuegos, PC, Tecnología y más", 1)
	page = strings.Replace(page, "{{ page.description }}", "Comunidad para curiosos, entusiastas y expertos de videojuegos, hardware y tecnología en general. Publica tus dudas y ayuda a los demás.", 2)
	page = strings.Replace(page, "{{ page.image }}", "http://spartangeek.com/images/default-post.jpg", 1)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, page)
}

func (this GeneralAPI) ByPass(c *gin.Context) {

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, this.Page)
}