package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"net/http"
	"html"
)

type ComponentAPI struct {
	Components *components.Module `inject:""`
	Page string
}

func (this ComponentAPI) Get(c *gin.Context) {
	
	slug := c.Param("slug")
	kind := c.Param("type")

	if kind == "tienda" {

		this.ByPass(c)
		return
	}

    if slug == "" {

		// Post not found, url hacked
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}
	
	component, err := this.Components.Get(bson.M{"slug": slug})

	if err != nil {

		// Post not found, url hacked
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	if kind != component.Type {

		var url string = "/componentes/" + component.Type + "/" + component.Slug
	    
		c.Redirect(http.StatusMovedPermanently, url)
		return
	}
    
    var name string = html.EscapeString(component.FullName)
    
    if name == "" {
        name = html.EscapeString(component.Name)    
    }
    
    var description string = html.EscapeString("Especificaciones y características del " + name + " y precios en tiendas. Comentarios de usuarios, ratings y preguntas.")
	var page string = this.Page

	page = strings.Replace(page, "SpartanGeek.com | Comunidad de tecnología, geeks y más", "SpartanGeek.com | Componentes | " + name , 1)
	page = strings.Replace(page, "{{ page.title }}", name, 1)
	page = strings.Replace(page, "{{ page.description }}", description, 2)

	if len(component.Image) > 0 {

		// First post image
		page = strings.Replace(page, "{{ page.image }}", "https://assets.spartangeek.com/components/" + component.Image, 1)

	} else {

		// Fallback to default image
		page = strings.Replace(page, "{{ page.image }}", "http://spartangeek.com/images/default-post.jpg", 1)
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, page)
}

func (this ComponentAPI) MigrateOld(c *gin.Context) {
	
	slug := c.Param("slug")

    if slug == "" {

		// Post not found, url hacked
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}
	
	component, err := this.Components.Get(bson.M{"slug": slug})

	if err != nil {

		// C not found, url hacked
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	var url string = "/componentes/" + component.Type + "/" + component.Slug
    
	c.Redirect(http.StatusMovedPermanently, url)
	return
}

func (this ComponentAPI) ByPass(c *gin.Context) {

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, this.Page)
}