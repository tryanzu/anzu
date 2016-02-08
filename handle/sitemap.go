package handle

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/gin-gonic/gin"
	"github.com/xuyu/goredis"
	"github.com/jinzhu/now"

	"time"
	"log"
)

type SitemapAPI struct {
	DataService  *mongo.Service `inject:""`
	CacheService *goredis.Redis `inject:""`
}

func (di *SitemapAPI) GetSitemap(c *gin.Context) {

	var urls []model.SitemapUrl
	var post model.Post
	var location string

	// Get the database interface from the DI
	database := di.DataService.Database

	iter := database.C("posts").Find(nil).Iter()

	for iter.Next(&post) {

		// Generate the post url
		location = "https://www.spartangeek.com/p/" + post.Slug + "/" + post.Id.Hex()

		// Add to the sitemap url
		urls = append(urls, model.SitemapUrl{Location: location, Updated: post.Updated.Format("2006-01-02T15:04:05.999999-07:00"), Priority: "0.6"})
	}

	urls = append(urls, model.SitemapUrl{Location: "http://www.spartangeek.com", Updated: time.Now().Format("2006-01-02T15:04:05.999999-07:00"), Priority: "1.0"})

	sitemap := model.SitemapSet{
		Urls:        urls,
		XMLNs:       "http://www.sitemaps.org/schemas/sitemap/0.9",
		XSI:         "http://www.w3.org/2001/XMLSchema-instance",
		XSILocation: "http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd",
	}

	c.XML(200, sitemap)
}

func (di *SitemapAPI) GetComponentsSitemap(c *gin.Context) {

	var urls []model.SitemapUrl
	var component components.ComponentModel
	var location string

	// Get the database interface from the DI
	database := di.DataService.Database

	iter := database.C("components").Find(nil).Limit(50000).Iter()
	
	// Let's say each component without store info changes once every month
	month := now.BeginningOfMonth()

	for iter.Next(&component) {

		// Generate the post url
		location = "https://www.spartangeek.com/componentes/" + component.Type + "/" + component.Slug
		updated := month.Format("2006-01-02T15:04:05.999999-07:00")

		if component.Activated {
			updated = component.Store.Updated.Format("2006-01-02T15:04:05.999999-07:00") 
		} 

		// Add to the sitemap url
		urls = append(urls, model.SitemapUrl{Location: location, Updated: updated, Priority: "0.9"})
	}

	log.Printf("count: %v\n", len(urls))

	urls = append(urls, model.SitemapUrl{Location: "http://www.spartangeek.com/componentes", Updated: month.Format("2006-01-02T15:04:05.999999-07:00"), Priority: "1.0"})
	urls = append(urls, model.SitemapUrl{Location: "http://www.spartangeek.com/componentes/tienda", Updated: month.Format("2006-01-02T15:04:05.999999-07:00"), Priority: "1.0"})

	sitemap := model.SitemapSet{
		Urls:        urls,
		XMLNs:       "http://www.sitemaps.org/schemas/sitemap/0.9",
		XSI:         "http://www.w3.org/2001/XMLSchema-instance",
		XSILocation: "http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd",
	}

	c.XML(200, sitemap)
}
