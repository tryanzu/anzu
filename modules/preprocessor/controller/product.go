package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"net/http"
	"html"
)

type ProductAPI struct {
	GCommerce *gcommerce.Module `inject:""`
	Page string
}

func (this ProductAPI) Legion(c *gin.Context) {
	
	slug := c.Param("slug")

    if slug == "" {

		// Post not found, url hacked
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	products := this.GCommerce.Products()
	product, err := products.GetByBson(bson.M{"slug": slug})

	if err != nil {

		// Post not found, url hacked
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}
    
    var name string = html.EscapeString(product.Name)
    var description string = html.EscapeString(name + " al mejor precio en México. Compra en Legión y aprovecha esta oferta por tiempo limitado.")
	var page string = this.Page

	page = strings.Replace(page, "SpartanGeek.com | Comunidad de tecnología, geeks y más", "SpartanGeek.com | Compra en Legión | " + name , 1)
	page = strings.Replace(page, "{{ page.title }}", name, 1)
	page = strings.Replace(page, "{{ page.description }}", description, 2)

	if len(product.Image) > 0 {

		// First post image
		page = strings.Replace(page, "{{ page.image }}", "https://assets.spartangeek.com/components/" + product.Image, 1)

	} else {

		// Fallback to default image
		page = strings.Replace(page, "{{ page.image }}", "http://spartangeek.com/images/default-post.jpg", 1)
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, page)
}