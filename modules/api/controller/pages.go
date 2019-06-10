package controller

import (
	"github.com/gin-gonic/gin"
	posts "github.com/tryanzu/core/board/posts"
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/helpers"
	"gopkg.in/mgo.v2/bson"
)

// HomePage is the site's entry point.
func HomePage(c *gin.Context) {
	conf := c.MustGet("config").(config.Anzu)
	title := conf.Site.Name
	if len(conf.Site.TitleMotto) > 0 {
		title += " - " + conf.Site.TitleMotto
	}
	c.HTML(200, "pages/index.tmpl", gin.H{
		"config":      conf,
		"title":       title,
		"description": c.MustGet("siteDescription").(string),
		"image":       c.MustGet("siteUrl").(string) + "/images/default-post.jpg",
	})
}

func UserPage(c *gin.Context) {
	id := bson.ObjectIdHex(c.Param("id"))
	usr, err := user.FindId(deps.Container, id)

	if err != nil {
		c.AbortWithStatus(404)
		return
	}

	c.HTML(200, "pages/index.tmpl", gin.H{
		"config":      c.MustGet("config").(config.Anzu),
		"title":       usr.UserName + " - Perfil de usuario",
		"description": "Explora las aportaciones y el perfil de " + usr.UserName + " en Buldar",
		"image":       c.MustGet("siteUrl").(string) + "/images/default-post.jpg",
	})
}

func PostPage(c *gin.Context) {
	id := bson.ObjectIdHex(c.Param("id"))
	post, err := posts.FindId(deps.Container, id)

	if err != nil {
		c.AbortWithStatus(404)
		return
	}

	if post.Slug != c.Param("slug") {
		c.Redirect(301, c.MustGet("siteUrl").(string)+"/p/"+post.Slug+"/"+post.Id.Hex())
		return
	}

	c.HTML(200, "pages/index.tmpl", gin.H{
		"config":      c.MustGet("config").(config.Anzu),
		"title":       post.Title,
		"description": helpers.Truncate(post.Content, 160),
		"image":       c.MustGet("siteUrl").(string) + "/images/default-post.jpg",
	})
}
