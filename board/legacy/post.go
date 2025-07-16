package handle

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olebedev/config"
	"github.com/tryanzu/core/board/legacy/model"
	posts "github.com/tryanzu/core/board/posts"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/acl"
	"github.com/tryanzu/core/modules/exceptions"
	"github.com/tryanzu/core/modules/feed"
	"github.com/tryanzu/core/modules/gaming"
	"github.com/xuyu/goredis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PostAPI struct {
	CacheService  *goredis.Redis               `inject:""`
	Feed          *feed.FeedModule             `inject:""`
	Errors        *exceptions.ExceptionsModule `inject:""`
	S3Bucket      *deps.S3Service              `inject:""`
	Gaming        *gaming.Module               `inject:""`
	ConfigService *config.Config               `inject:""`
	Acl           *acl.Module                  `inject:""`
}

func (di PostAPI) FeedGet(c *gin.Context) {
	var (
		feed     []model.FeedPost
		search   = bson.M{}
		offset   = 0
		limit    = 10
		database = deps.Container.Mgo()
	)

	if n, err := strconv.Atoi(c.Query("offset")); err == nil && n > 0 {
		offset = n
	}

	if n, err := strconv.Atoi(c.Query("limit")); err == nil && n < 40 {
		limit = n
	}

	if s := c.Query("search"); len(s) > 0 {
		search["$text"] = bson.M{"$search": s}
	}

	if id := c.Query("category"); primitive.IsValidObjectID(id) {
		categoryID, _ := primitive.ObjectIDFromHex(id)
		search["category"] = categoryID
	}

	if slug := c.Query("category"); len(slug) > 0 && !primitive.IsValidObjectID(slug) {
		var category struct {
			ID primitive.ObjectID `bson:"_id,omitempty"`
		}

		if err := database.Collection("categories").FindOne(context.Background(), bson.M{"slug": slug}, nil).Decode(&category); err == nil {
			search["category"] = category.ID
		}
	}

	if t, err := time.Parse(time.RFC3339Nano, c.Query("after")); len(c.Query("after")) > 0 && err == nil {
		search["created_at"] = bson.M{"$lt": t}
	}

	if t, err := time.Parse(time.RFC3339Nano, c.Query("before")); len(c.Query("before")) > 0 && err == nil {
		search["created_at"] = bson.M{"$gt": t}
	}

	relevant := c.Query("relevant")
	user_order := false
	count := 0

	if author := c.Query("user_id"); len(author) > 0 && primitive.IsValidObjectID(author) {
		authorID, _ := primitive.ObjectIDFromHex(author)
		search["user_id"] = authorID
		user_order = true
	}

	if categories := c.Query("categories"); len(categories) > 0 {
		var within []primitive.ObjectID
		list := strings.Split(categories, ",")
		for _, cid := range list {
			if primitive.IsValidObjectID(cid) {
				categoryID, _ := primitive.ObjectIDFromHex(cid)
				within = append(within, categoryID)
			}
		}

		if len(within) > 0 {
			search["category"] = bson.M{"$in": within}
		}
	}

	if relevant != "" {
		list, err := posts.FindRateList(deps.Container, relevant, offset*4, limit*4)
		if err != nil {
			log.Printf("[err] %v\n", err)
		}
		if err == nil && len(list) > 0 {
			var temp []model.FeedPost
			cursor, err := database.Collection("posts").Find(context.Background(), bson.M{
				"_id":        bson.M{"$in": list},
				"deleted_at": bson.M{"$exists": false},
				"created_at": bson.M{"$gte": time.Now().Add(time.Hour * 24 * 30 * -1)},
			}, nil)
			if err != nil {
				panic(err)
			}
			defer cursor.Close(context.Background())
			err = cursor.All(context.Background(), &temp)

			if err != nil {
				panic(err)
			}

			feed = []model.FeedPost{}

			// Using the temp feed we will have to manually order them by the natural order given by the relevant list
			for _, id := range list {
				for _, post := range temp {
					if post.Id == id {
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
		search["deleted_at"] = bson.M{"$exists": false}

		// Prepare the database to fetch the feed
		var cursor *mongo.Cursor
		var err error

		// Add the sort depending on the context
		if user_order {
			countResult, _ := database.Collection("posts").CountDocuments(context.Background(), search)
			count = int(countResult)
			cursor, err = database.Collection("posts").Find(context.Background(), search, nil)
		} else {
			cursor, err = database.Collection("posts").Find(context.Background(), search, nil)
		}

		if err != nil {
			panic(err)
		}
		defer cursor.Close(context.Background())

		// Get the results from the feed algo
		err = cursor.All(context.Background(), &feed)
		if err != nil {
			panic(err)
		}
	}

	var authors []primitive.ObjectID
	var list []primitive.ObjectID
	var users []model.User

	for _, post := range feed {
		list = append(list, post.Id)
		authors = append(authors, post.UserId)
	}

	if _, signed := c.Get("user_id"); signed {
		events.In <- events.PostsReached(signs(c), list)
	}

	// Get the users needed by the feed
	if len(authors) > 0 {
		cursor, err := database.Collection("users").Find(context.Background(), bson.M{"_id": bson.M{"$in": authors}}, nil)
		if err != nil {
			panic(err)
		}
		defer cursor.Close(context.Background())

		err = cursor.All(context.Background(), &users)
		if err != nil {
			panic(err)
		}
	}

	if len(feed) > 0 {

		usersMap := make(map[primitive.ObjectID]model.User)

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
					Seen:     postUser.Seen,
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

func signs(c *gin.Context) events.UserSign {
	usr := c.MustGet("userID").(primitive.ObjectID)
	sign := events.UserSign{
		UserID: usr,
	}
	if r := c.Query("reason"); len(r) > 0 {
		sign.Reason = r
	}
	return sign
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
		name = primitive.NewObjectID().Hex()

		if extension == "" {

			extension = ".jpg"
		}

		path := "posts/" + name + extension
		err = di.S3Bucket.PutObject(path, data, dataType)

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
	// Get the post using the id
	id := c.Params.ByName("id")
	if !primitive.IsValidObjectID(id) {
		c.JSON(400, gin.H{
			"message": "Invalid request, no valid params.",
			"status":  "error",
		})
		return
	}

	// Get the post using the slug
	user_id := c.MustGet("user_id")
	uid, _ := primitive.ObjectIDFromHex(user_id.(string))
	bson_id, _ := primitive.ObjectIDFromHex(id)
	post, err := di.Feed.Post(bson_id)

	if err != nil {
		c.JSON(404, gin.H{
			"message": "Couldnt find the post",
			"status":  "error",
		})
		return
	}

	user := di.Acl.User(uid)
	if user.CanDeletePost(post) == false {
		c.JSON(400, gin.H{"message": "Can't delete others posts.", "status": "error"})
		return
	}

	_, err = deps.Container.Mgo().Collection("posts").UpdateOne(context.Background(), bson.M{"_id": post.Id}, bson.M{
		"$set":   bson.M{"deleted": true, "deleted_at": time.Now()},
		"$unset": bson.M{"pinned": ""},
	})
	if err != nil {
		panic(err)
	}

	events.In <- events.DeletePost(signs(c), post.Id)

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

	events.In <- events.RawEmit("feed", "action", params)
}

func (di PostAPI) downloadAssetFromUrl(from string, post_id primitive.ObjectID) error {

	// Recover from any panic even inside this goroutine
	defer di.Errors.Recover()

	// Get the database interface from the DI
	database := deps.Container.Mgo()
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
		name = primitive.NewObjectID().Hex()

		if extension != "" {

			name = name + extension
		} else {

			// If no extension is provided on the url then add a dummy one
			name = name + ".jpg"
		}

		path := "posts/" + name
		err = di.S3Bucket.PutObject(path, data, dataType)

		if err != nil {

			panic(err)
		}

		var post model.Post

		err = database.Collection("posts").FindOne(context.Background(), bson.M{"_id": post_id}).Decode(&post)

		if err == nil {

			post_content := post.Content

			// Replace the url on the comment
			if strings.Contains(post_content, from) {

				content := strings.Replace(post_content, from, amazon_url+path, -1)

				// Update the comment
				deps.Container.Mgo().Collection("posts").UpdateOne(context.Background(), bson.M{"_id": post_id}, bson.M{"$set": bson.M{"content": content}})
			}

		}
	}

	response.Body.Close()

	return nil
}

func (di PostAPI) resetUserCategoryCounter(category string, user_id primitive.ObjectID) {

	// Recover from any panic even inside this goroutine
	defer di.Errors.Recover()

	// Replace the slug dash with underscore
	counter := strings.Replace(category, "-", "_", -1)
	find := "counters." + counter + ".counter"
	updated_at := "counters." + counter + ".updated_at"

	// Update the collection of counters
	_, err := deps.Container.Mgo().Collection("counters").UpdateOne(context.Background(), bson.M{"user_id": user_id}, bson.M{"$set": bson.M{find: 0, updated_at: time.Now()}})

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
	deps.Container.Mgo().Collection("counters").UpdateMany(context.Background(), bson.M{}, bson.M{"$inc": bson.M{find: 1}})

	return
}
