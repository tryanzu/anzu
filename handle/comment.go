package handle

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/cosn/firebase"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/notifications"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/ftrvxmtrx/gravatar"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/kennygrant/sanitize"
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"gopkg.in/mgo.v2/bson"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CommentAPI struct {
	DataService   *mongo.Service                     `inject:""`
	Firebase      *firebase.Client                   `inject:""`
	ConfigService *config.Config                     `inject:""`
	S3Bucket      *s3.Bucket                         `inject:""`
	Notifications *notifications.NotificationsModule `inject:""`
	Errors        *exceptions.ExceptionsModule       `inject:""`
	Gaming        *GamingAPI                         `inject:""`
}

func (di *CommentAPI) CommentAdd(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	id := c.Params.ByName("id")

	if bson.IsObjectIdHex(id) == false {

		c.JSON(400, gin.H{"error": "Invalid request, no valid params.", "status": 701})
		return
	}

	user_id := c.MustGet("user_id")
	user_bson_id := bson.ObjectIdHex(user_id.(string))

	var comment model.CommentForm

	if c.BindWith(&comment, binding.JSON) == nil {

		// Get the post using the slug
		id := bson.ObjectIdHex(id)

		var post model.Post

		err := database.C("posts").FindId(id).One(&post)

		if err != nil {

			c.JSON(404, gin.H{"message": "Couldnt find the post", "status": "error"})
			return
		}

		if post.NoComments == true {

			c.JSON(403, gin.H{"status": "error", "message": "Commnets now allowed at all."})
			return
		}

		votes := model.Votes{
			Up:   0,
			Down: 0,
		}

		// Html sanitize
		content := sanitize.HTML(comment.Content)
		comment := model.Comment{
			UserId:  user_bson_id,
			Votes:   votes,
			Content: content,
			Created: time.Now(),
		}

		urls, _ := regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)

		var assets []string

		assets = urls.FindAllString(content, -1)

		for _, asset := range assets {

			// Download the asset on other routine in order to non block the API request
			go di.downloadAssetFromUrl(asset, post.Id)
		}

		// Update the post and push the comments
		change := bson.M{"$push": bson.M{"comments.set": comment}, "$set": bson.M{"updated_at": time.Now()}, "$inc": bson.M{"comments.count": 1}}
		err = database.C("posts").Update(bson.M{"_id": post.Id}, change)

		if err != nil {
			panic(err)
		}

		position := strconv.Itoa(len(post.Comments.Set))

		// Process the mentions. TODO - Determine race conditions
		go di.Notifications.ParseContentMentions(notifications.MentionParseObject{
			Type:          "comment",
			RelatedNested: position,
			Content:       comment.Content,
			Title:         post.Title,
			Author:        user_bson_id,
			Post:          post,
		})

		// Check if we need to add participant
		users := post.Users
		need_add := true

		for _, already_within := range users {

			if already_within == user_bson_id {

				need_add = false
			}
		}

		if need_add == true {

			// Add the user to the user list
			change := bson.M{"$push": bson.M{"users": user_bson_id}}
			err = database.C("posts").Update(bson.M{"_id": post.Id}, change)

			if err != nil {
				panic(err)
			}
		}

		// Triggers when the author of the comment is not the post's author
		if post.UserId != user_bson_id {

			// Notify the post's author
			go di.notifyCommentPostAuth(post, user_bson_id)

			// Add the gamification contribution
			go di.Gaming.Related(user_bson_id).Did("comment")
		}

		c.JSON(200, gin.H{"status": "okay", "message": comment.Content, "position": position})
		return
	}

	c.JSON(401, gin.H{"error": "Not authorized.", "status": 704})
}

func (di *CommentAPI) CommentUpdate(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	id := c.Params.ByName("id")
	index := c.Params.ByName("index")

	if bson.IsObjectIdHex(id) == false {

		c.JSON(400, gin.H{"error": "Invalid request, no valid params.", "status": 701})
		return
	}

	user_id := c.MustGet("user_id")
	user_bson_id := bson.ObjectIdHex(user_id.(string))

	var commentForm model.CommentForm
	var post model.Post
	var assets []string

	if c.BindWith(&commentForm, binding.JSON) == nil {

		// Get the post using the slug
		id := bson.ObjectIdHex(id)
		err := database.C("posts").FindId(id).One(&post)

		if err != nil {

			c.JSON(404, gin.H{"message": "Couldnt find the post", "status": "error"})
			return
		}

		comment_index, err := strconv.Atoi(index)

		if err != nil || len(post.Comments.Set)-1 < comment_index {

			c.JSON(400, gin.H{"message": "Invalid request, no valid comment index.", "status": "error"})
			return
		}

		comment := post.Comments.Set[comment_index]
		content := html.EscapeString(commentForm.Content)

		if user_bson_id != comment.UserId {

			c.JSON(400, gin.H{"message": "Can't edit others comments.", "status": "error"})
			return
		}

		// Get assets from the new content
		urls, _ := regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
		assets = urls.FindAllString(content, -1)

		for _, asset := range assets {

			// Download the asset on other routine in order not block the API request
			go di.downloadAssetFromUrl(asset, post.Id)
		}

		// Database post comment path
		comment_path := "comments.set." + index + ".content"
		comment_path_updated := "comments.set." + index + ".updated_at"

		// Update the post and push the comments
		err = database.C("posts").Update(bson.M{"_id": post.Id}, bson.M{"$set": bson.M{comment_path: content, "updated_at": time.Now(), comment_path_updated: time.Now()}})

		if err != nil {
			panic(err)
		}

		// Process the mentions. TODO - Determine race conditions
		go di.Notifications.ParseContentMentions(notifications.MentionParseObject{
			Type:          "comment",
			RelatedNested: index,
			Content:       content,
			Title:         post.Title,
			Author:        user_bson_id,
			Post:          post,
		})

		c.JSON(200, gin.H{"status": "okay", "message": comment.Content})
		return
	}

	c.JSON(401, gin.H{"error": "Not authorized.", "status": 704})
}

func (di *CommentAPI) CommentDelete(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	id := c.Params.ByName("id")
	index := c.Params.ByName("index")

	if bson.IsObjectIdHex(id) == false {

		c.JSON(400, gin.H{"error": "Invalid request, no valid params.", "status": 701})
		return
	}

	user_id := c.MustGet("user_id")
	user_bson_id := bson.ObjectIdHex(user_id.(string))

	var post model.Post

	// Get the post using the slug
	bson_id := bson.ObjectIdHex(id)
	err := database.C("posts").FindId(bson_id).One(&post)

	if err != nil {

		c.JSON(404, gin.H{"message": "Couldnt find the post", "status": "error"})
		return
	}

	comment_index, err := strconv.Atoi(index)

	if err != nil || len(post.Comments.Set)-1 < comment_index {

		c.JSON(400, gin.H{"message": "Invalid request, no valid comment index.", "status": "error"})
		return
	}

	comment := post.Comments.Set[comment_index]

	if user_bson_id != comment.UserId {

		c.JSON(400, gin.H{"message": "Can't delete others comments.", "status": "error"})
		return
	}

	// Database post comment path
	comment_path := "comments.set." + index + ".deleted_at"

	// Update the post and push the comments
	err = database.C("posts").Update(bson.M{"_id": post.Id}, bson.M{"$set": bson.M{comment_path: time.Now()}, "$inc": bson.M{"comments.count": -1}})

	if err != nil {
		panic(err)
	}

	c.JSON(200, gin.H{"status": "okay"})
	return
} 


func (di *CommentAPI) notifyCommentPostAuth(post model.Post, user_id bson.ObjectId) {

	// Recover from any panic even inside this goroutine
	defer di.Errors.Recover()

	// Get the comment author
	var user model.User
	var notifications model.UserFirebaseNotifications

	err := di.DataService.Database.C("users").Find(bson.M{"_id": user_id}).One(&user)

	if err == nil {

		// Gravatar url
		emailHash := gravatar.EmailHash(user.Email)
		image := gravatar.GetAvatarURL("http", emailHash, "http://spartangeek.com/images/default-avatar.png", 80)

		// Construct the notification message
		title := fmt.Sprintf("Nuevo comentario de **%s**", user.UserName)
		message := post.Title

		// Process notification using firebase
		authorPath := "users/" + post.UserId.Hex() + "/notifications"
		authorRef := di.Firebase.Child(authorPath, nil, &notifications)

		authorRef.Set("count", notifications.Count+1, nil)

		notification := &model.UserFirebaseNotification{
			UserId:       user_id,
			RelatedId:    post.Id,
			RelatedExtra: post.Slug,
			Position:     post.Comments.Count,
			Title:        title,
			Text:         message,
			Related:      "comment",
			Seen:         false,
			Image:        image.String(),
			Created:      time.Now(),
			Updated:      time.Now(),
		}

		authorRef.Child("list", nil, nil).Push(notification, nil)
	}
}

func (di *CommentAPI) downloadAssetFromUrl(from string, post_id bson.ObjectId) error {

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

			for index := range post.Comments.Set {

				comment := post.Comments.Set[index].Content

				// Replace the url on the comment
				if strings.Contains(comment, from) {

					var rem bytes.Buffer

					// Make the push string
					rem.WriteString("comments.set.")
					rem.WriteString(strconv.Itoa(index))
					rem.WriteString(".content")

					ctc := rem.String()

					content := strings.Replace(comment, from, amazon_url+path, -1)

					// Update the comment
					di.DataService.Database.C("posts").Update(bson.M{"_id": post_id}, bson.M{"$set": bson.M{ctc: content}})
				}
			}
		}
	}

	response.Body.Close()

	return nil
}
