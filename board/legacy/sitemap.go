package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/deps"
	"github.com/xuyu/goredis"


	"time"
)

type SitemapAPI struct {
	CacheService *goredis.Redis `inject:""`
}

func (di *SitemapAPI) GetSitemap(c *gin.Context) {

	var urls []model.SitemapUrl
	var post model.Post
	var location string

	// Get the database interface from the DI
	database := deps.Container.Mgo()
	count := 1000
	iter := database.C("posts").Find(nil).Sort("-$natural").Limit(count).Iter()

	legacy := deps.Container.Config()
	siteurl, err := legacy.String("site.url")
	if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "site.url not found in config"})
			return
		}
	for iter.Next(&post) {
		// Generate the post url
		location = siteurl+"/p/" + post.Slug + "/" + post.Id.Hex() + "/"

		// Add to the sitemap url
		urls = append(urls, model.SitemapUrl{Location: location, Updated: post.Updated.Format("2006-01-02T15:04:05.999999-07:00"), Priority: "0.6"})
	}

	urls = append(urls, model.SitemapUrl{Location: siteurl, Updated: time.Now().Format("2006-01-02T15:04:05.999999-07:00"), Priority: "1.0"})

	sitemap := model.SitemapSet{
		Urls:        urls,
		XMLNs:       "http://www.sitemaps.org/schemas/sitemap/0.9",
		XSI:         "http://www.w3.org/2001/XMLSchema-instance",
		XSILocation: "http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd",
	}
	c.XML(200, sitemap)
}
