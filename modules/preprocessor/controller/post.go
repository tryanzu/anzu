package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"regexp"
	"strings"
)

type PostAPI struct {
	Feed *feed.FeedModule `inject:""`
	Page string
}

func (this *PostAPI) Get(c *gin.Context) {

	post_id := c.Param("id")
	slug := c.Param("slug")

	if bson.IsObjectIdHex(post_id) == false {

		// Invalid post id, url hacked
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	id := bson.ObjectIdHex(post_id)
	post, err := this.Feed.Post(id)

	if err != nil {

		// Post not found, url hacked
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	post_data := post.Data()

	if slug != post_data.Slug {

		// Invalid slug, redirect to correct one permanently
		c.Redirect(http.StatusMovedPermanently, "/p/"+post_data.Slug+"/"+post_data.Id.Hex())
		return
	}

	var description string = truncate(post_data.Content, 155) + "..."
	var page string = this.Page

	page = strings.Replace(page, "Buldar | Comunidad de tecnología, geeks y más", "SpartanGeek.com | "+post_data.Title, 1)
	page = strings.Replace(page, "{{ page.title }}", post_data.Title, 1)
	page = strings.Replace(page, "{{ page.description }}", description, 2)

	// Replace post image
	urls, _ := regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)

	var assets []string
	assets = urls.FindAllString(post_data.Content, -1)

	if len(assets) > 0 {

		// First post image
		page = strings.Replace(page, "{{ page.image }}", assets[0], 1)

	} else {

		// Fallback to default image
		page = strings.Replace(page, "{{ page.image }}", "https://buldar.com/images/default-post.jpg", 1)
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, page)
}

func (this *PostAPI) ByPass(c *gin.Context) {

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, this.Page)
}

func truncate(s string, length int) string {
	var numRunes = 0
	for index, _ := range s {
		numRunes++
		if numRunes > length {
			return s[:index]
		}
	}
	return s
}
