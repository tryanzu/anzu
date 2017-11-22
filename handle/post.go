package handle

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type PostAPI struct {
	DataService   *mongo.Service               `inject:""`
	CacheService  *goredis.Redis               `inject:""`
	Feed          *feed.FeedModule             `inject:""`
	Collector     CollectorAPI                 `inject:"inline"`
	Errors        *exceptions.ExceptionsModule `inject:""`
	S3Bucket      *s3.Bucket                   `inject:""`
	Gaming        *gaming.Module               `inject:""`
	ConfigService *config.Config               `inject:""`
	Acl           *acl.Module                  `inject:""`
}

/**
 * Avaliable components
 *
 * List of valid components along a recommendation post
 */
var avaliable_components = []string{
	"cpu", "motherboard", "ram", "cabinet", "screen", "storage", "cooler", "power", "videocard",
}

func (di PostAPI) FeedGet(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database
	redis := di.CacheService

	var feed []model.FeedPost
	offset := 0
	limit := 10

	o := c.Query("offset")
	l := c.Query("limit")
	f := c.Query("category")
	relevant := c.Query("relevant")
	fltr_categories := c.Query("categories")
	before := c.Query("before")
	after := c.Query("after")
	from_author := c.Query("user_id")
	user_id, signed_in := c.Get("user_id")

	// Check if offset has been specified
	if o != "" {
		off, err := strconv.Atoi(o)

		if err != nil || off < 0 {
			c.JSON(401, gin.H{"message": "Invalid request, check params.", "status": "error", "code": 901})
			return
		}

		offset = off
	}

	// Check if limit has been specified
	if l != "" {

		lim, err := strconv.Atoi(l)

		if err != nil || lim <= 0 || lim > 40 {
			c.JSON(401, gin.H{"message": "Invalid request, check params.", "status": "error", "code": 901})
			return
		}

		limit = lim
	}

	search := make(bson.M)
	if bson.IsObjectIdHex(f) {
		search["category"] = bson.ObjectIdHex(f)
	}

	if t, err := time.Parse(time.RFC3339Nano, after); after != "" && err == nil {
		search["created_at"] = bson.M{"$lt": t}
	}

	if t, err := time.Parse(time.RFC3339Nano, before); before != "" && err == nil {
		search["created_at"] = bson.M{"$gt": t}
	}

	user_order := false
	count := 0

	if from_author != "" && bson.IsObjectIdHex(from_author) {

		search["user_id"] = bson.ObjectIdHex(from_author)

		user_order = true
	}

	_, filter_by_category := search["category"]

	// Get the list of categories a user is following when the request is authenticated
	if signed_in && !filter_by_category && !user_order {

		var user_categories []bson.ObjectId

		user_categories_list, err := redis.SMembers("user:categories:" + user_id.(string))

		if err != nil {
			panic(err)
		}

		if len(user_categories_list) == 0 {

			var user model.User

			err := database.C("users").Find(bson.M{"_id": bson.ObjectIdHex(user_id.(string))}).One(&user)

			if err != nil {
				panic(err)
			}

			if len(user.Categories) > 0 {

				var category_members []string

				for _, category_id := range user.Categories {

					user_categories = append(user_categories, category_id)
					category_members = append(category_members, category_id.Hex())
				}

				// Create the set inside redis and move on
				redis.SAdd("user:categories:"+user_id.(string), category_members...)
			}

		} else {

			for _, category_id := range user_categories_list {

				if bson.IsObjectIdHex(category_id) {

					user_categories = append(user_categories, bson.ObjectIdHex(category_id))
				}
			}
		}

		if len(user_categories) > 0 {
			search["category"] = bson.M{"$in": user_categories}
		}

	} else if fltr_categories != "" {

		var user_categories []bson.ObjectId

		provided_categories := strings.Split(fltr_categories, ",")

		for _, category_id := range provided_categories {

			if bson.IsObjectIdHex(category_id) {

				user_categories = append(user_categories, bson.ObjectIdHex(category_id))
			}
		}

		if len(user_categories) > 0 {

			search["category"] = bson.M{"$in": user_categories}
		}
	}

	if relevant != "" {

		// Calculate the offset using the limit
		list_start := offset
		list_end := offset + limit

		relevant_date := relevant

		relevant_list, err := redis.ZRevRange("feed:relevant:"+relevant_date, list_start, list_end, false)

		if err == nil && len(relevant_list) > 0 {

			var temp_feed []model.FeedPost
			var relevant_ids []bson.ObjectId

			for _, relevant_id := range relevant_list {

				relevant_ids = append(relevant_ids, bson.ObjectIdHex(relevant_id))
			}

			err := database.C("posts").Find(bson.M{"_id": bson.M{"$in": relevant_ids}}).Select(bson.M{"comments.set": 0, "content": 0, "components": 0}).All(&temp_feed)

			if err != nil {
				panic(err)
			}

			feed = []model.FeedPost{}

			// Using the temp feed we will have to manually order them by the natural order given by the relevant list
			for _, relevant_id := range relevant_ids {

				for _, post := range temp_feed {

					if post.Id == relevant_id {

						feed = append(feed, post)
						break
					}
				}
			}

		} else {

			c.JSON(200, gin.H{"feed": []string{}, "offset": offset, "limit": limit})

			return
		}

	} else {

		// Get all but deleted
		if signed_in {
			search["$or"] = []bson.M{
				{"deleted_at": bson.M{"$exists": false}},
				{"user_id": bson.ObjectIdHex(user_id.(string))},
			}
		} else {
			search["deleted_at"] = bson.M{"$exists": false}
		}

		// Prepare the database to fetch the feed
		posts_collection := database.C("posts")
		get_feed := posts_collection.Find(search).Select(bson.M{"comments.set": 0, "content": 0, "components": 0})

		// Add the sort depending on the context
		if user_order {
			count, _ = get_feed.Count()
			get_feed = get_feed.Sort("-created_at")
		} else {

			get_feed = get_feed.Sort("-pinned", "-created_at")
		}

		// Add the limits of the resultset
		get_feed = get_feed.Limit(limit).Skip(offset)

		// Get the results from the feed algo
		err := get_feed.All(&feed)

		if err != nil {
			panic(err)
		}
	}

	var authors []bson.ObjectId
	var list []bson.ObjectId
	var users []model.User

	for _, post := range feed {
		list = append(list, post.Id)
		authors = append(authors, post.UserId)
	}

	if signed_in {

		// Save the activity in other routine
		go di.Collector.Activity(model.Activity{UserId: bson.ObjectIdHex(user_id.(string)), Event: "feed", List: list})
	}

	// Update the feed rates for the most important stuff
	go di.Feed.UpdateFeedRates(feed)

	// Get the users needed by the feed
	err := database.C("users").Find(bson.M{"_id": bson.M{"$in": authors}}).All(&users)

	if err != nil {
		panic(err)
	}

	if len(feed) > 0 {

		usersMap := make(map[bson.ObjectId]model.User)

		for _, user := range users {

			usersMap[user.Id] = user
		}

		for index := range feed {

			post := &feed[index]

			if _, okay := usersMap[post.UserId]; okay {

				postUser := usersMap[post.UserId]

				var authorLevel int

				if _, stepOkay := postUser.Profile["step"]; stepOkay {
					authorLevel = 1
				} else {
					authorLevel = 0
				}

				post.Author = model.User{
					Id:       postUser.Id,
					UserName: postUser.UserName,
					Step:     authorLevel,
					Image:    postUser.Image,
				}
			}
		}

		if count > 0 {

			c.JSON(200, gin.H{"feed": feed, "offset": offset, "limit": limit, "count": count})
		} else {

			c.JSON(200, gin.H{"feed": feed, "offset": offset, "limit": limit})
		}

	} else {

		c.JSON(200, gin.H{"feed": []string{}, "offset": offset, "limit": limit})
	}
}

func (di PostAPI) GetLightweight(c *gin.Context) {

	// Get the post ID
	id := c.Params.ByName("id")

	if bson.IsObjectIdHex(id) == false {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid user id."})
		return
	}

	post_id := bson.ObjectIdHex(id)
	post, err := di.Feed.Post(post_id)

	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	data := post.Data()
	data.Content = html.UnescapeString(data.Content)

	c.JSON(200, data)
}

func (di PostAPI) PostUploadAttachment(c *gin.Context) {

	// Check the file inside the request
	file, header, err := c.Request.FormFile("file")

	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Could not get the file..."})
		return
	}

	defer file.Close()

	// Read all the bytes from the image
	data, err := ioutil.ReadAll(file)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Could not read the file contents..."})
		return
	}

	// Detect the downloaded file type
	dataType := http.DetectContentType(data)

	if dataType[0:5] == "image" {

		var extension, name string

		extension = filepath.Ext(header.Filename)
		name = bson.NewObjectId().Hex()

		if extension == "" {

			extension = ".jpg"
		}

		path := "posts/" + name + extension
		err = di.S3Bucket.Put(path, data, dataType, s3.ACL("public-read"))

		if err != nil {
			panic(err)
		}

		s3_url := "https://s3-us-west-1.amazonaws.com/spartan-board/" + path

		// Done
		c.JSON(200, gin.H{"status": "okay", "url": s3_url})

		return
	}

	c.JSON(400, gin.H{"status": "error", "message": "Could not detect an image file..."})
}

func (di PostAPI) PostDelete(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	// Get the post using the id
	id := c.Params.ByName("id")

	if bson.IsObjectIdHex(id) == false {

		c.JSON(400, gin.H{"message": "Invalid request, no valid params.", "status": "error"})
		return
	}

	// Get the post using the slug
	user_id := c.MustGet("user_id")
	user_bson_id := bson.ObjectIdHex(user_id.(string))
	bson_id := bson.ObjectIdHex(id)
	post, err := di.Feed.Post(bson_id)

	if err != nil {

		c.JSON(404, gin.H{"message": "Couldnt find the post", "status": "error"})
		return
	}

	user := di.Acl.User(user_bson_id)

	if user.CanDeletePost(post) == false {

		c.JSON(400, gin.H{"message": "Can't delete others posts.", "status": "error"})
		return
	}

	err = database.C("posts").Update(bson.M{"_id": post.Id}, bson.M{"$set": bson.M{"deleted": true, "deleted_at": time.Now()}, "$unset": bson.M{"pinned": ""}})

	if err != nil {
		panic(err)
	}

	go deps.Container.Transmit().Emit("feed", "action", map[string]interface{}{
		"fire": "delete-post",
		"id":   post.Id.Hex(),
	})

	c.JSON(200, gin.H{"status": "okay"})
}

func (di PostAPI) syncUsersFeed(post *model.Post) {
	defer di.Errors.Recover()

	params := map[string]interface{}{
		"fire":     "new-post",
		"category": post.Category.Hex(),
		"user_id":  post.UserId.Hex(),
		"id":       post.Id.Hex(),
		"slug":     post.Slug,
	}

	deps.Container.Transmit().Emit("feed", "action", params)
}

func (di PostAPI) downloadAssetFromUrl(from string, post_id bson.ObjectId) error {

	// Recover from any panic even inside this goroutine
	defer di.Errors.Recover()

	// Get the database interface from the DI
	database := di.DataService.Database
	amazon_url, err := di.ConfigService.String("amazon.url")

	if err != nil {
		panic(err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Download the file
	response, err := client.Get(from)
	if err != nil {
		return errors.New(fmt.Sprint("Error while downloading", from, "-", err))
	}

	// Read all the bytes to the image
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.New(fmt.Sprint("Error while downloading", from, "-", err))
	}

	// Detect the downloaded file type
	dataType := http.DetectContentType(data)

	if dataType[0:5] == "image" {

		var extension, name string

		// Parse the filename
		u, err := url.Parse(from)

		if err != nil {
			return errors.New(fmt.Sprint("Error while parsing url", from, "-", err))
		}

		extension = filepath.Ext(u.Path)
		name = bson.NewObjectId().Hex()

		if extension != "" {

			name = name + extension
		} else {

			// If no extension is provided on the url then add a dummy one
			name = name + ".jpg"
		}

		path := "posts/" + name
		err = di.S3Bucket.Put(path, data, dataType, s3.ACL("public-read"))

		if err != nil {

			panic(err)
		}

		var post model.Post

		err = database.C("posts").Find(bson.M{"_id": post_id}).One(&post)

		if err == nil {

			post_content := post.Content

			// Replace the url on the comment
			if strings.Contains(post_content, from) {

				content := strings.Replace(post_content, from, amazon_url+path, -1)

				// Update the comment
				di.DataService.Database.C("posts").Update(bson.M{"_id": post_id}, bson.M{"$set": bson.M{"content": content}})
			}

		}
	}

	response.Body.Close()

	return nil
}

func (di PostAPI) resetUserCategoryCounter(category string, user_id bson.ObjectId) {

	// Recover from any panic even inside this goroutine
	defer di.Errors.Recover()

	// Replace the slug dash with underscore
	counter := strings.Replace(category, "-", "_", -1)
	find := "counters." + counter + ".counter"
	updated_at := "counters." + counter + ".updated_at"

	// Update the collection of counters
	err := di.DataService.Database.C("counters").Update(bson.M{"user_id": user_id}, bson.M{"$set": bson.M{find: 0, updated_at: time.Now()}})

	if err != nil {
		panic(err)
	}

	return
}

func (di PostAPI) addUserCategoryCounter(category string) {

	// Recover from any panic even inside this goroutine
	defer di.Errors.Recover()

	// Replace the slug dash with underscore
	counter := strings.Replace(category, "-", "_", -1)
	find := "counters." + counter + ".counter"

	// Update the collection of counters
	di.DataService.Database.C("counters").UpdateAll(nil, bson.M{"$inc": bson.M{find: 1}})

	return
}
