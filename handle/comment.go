package handle

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/cosn/firebase"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/ftrvxmtrx/gravatar"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/mitchellh/goamz/s3"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/url"
	"regexp"
	"io/ioutil"
	"strings"
	"time"
	"mime"
)

type CommentAPI struct {
	DataService *mongo.Service   `inject:""`
	Firebase    *firebase.Client `inject:""`
	S3Bucket    *s3.Bucket      `inject:""`
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

		// Posts collection
		collection := database.C("posts")

		var post model.Post

		err := collection.FindId(id).One(&post)

		if err != nil {

			c.JSON(404, gin.H{"error": "Couldnt find the post", "status": 705})
			return
		}

		votes := model.Votes{
			Up:   0,
			Down: 0,
		}

		comment := model.Comment{
			UserId:  user_bson_id,
			Votes:   votes,
			Content: comment.Content,
			Created: time.Now(),
		}

		urls, _ := regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
		mentions, _ := regexp.Compile(`(?i)\B\@([\w\-]+)`)

		var assets []string

		assets = urls.FindAllString(comment.Content, -1)

		for _, asset := range assets {

			// Download the asset on other routine in order to non block the API request
			go di.downloadAssetFromUrl(asset)
		}

		var mentions_users []string

		mentions_users = mentions.FindAllString(comment.Content, -1)

		for _, mention_user := range mentions_users {

			go di.mentionUserComment(mention_user, post, user_bson_id)

			// Replace the mentioned user
			markdown := "[" + mention_user + "](/u/" + mention_user[1:] + " \"" + mention_user[1:] + "\")"
			comment.Content = strings.Replace(comment.Content, mention_user, markdown, -1)
		}

		// Update the post and push the comments
		change := bson.M{"$push": bson.M{"comments.set": comment}, "$set": bson.M{"updated_at": time.Now()}, "$inc": bson.M{"comments.count": 1}}
		err = collection.Update(bson.M{"_id": post.Id}, change)

		if err != nil {
			panic(err)
		}

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
			err = collection.Update(bson.M{"_id": post.Id}, change)

			if err != nil {
				panic(err)
			}
		}

		// Notifications for the author
		if post.UserId != user_bson_id {

			go di.notifyCommentPostAuth(post, user_bson_id)
		}

		c.JSON(200, gin.H{"message": "okay", "status": 706})
		return
	}

	c.JSON(401, gin.H{"error": "Not authorized.", "status": 704})
}

func (di *CommentAPI) notifyCommentPostAuth(post model.Post, user_id bson.ObjectId) {

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
			UserId:    user_id,
			RelatedId: post.Id,
			RelatedExtra: post.Slug,
			Position:  post.Comments.Count,
			Title:     title,
			Text:      message,
			Related:   "comment",
			Seen:      false,
			Image:     image.String(),
			Created:   time.Now(),
			Updated:   time.Now(),
		}

		authorRef.Child("list", nil, nil).Push(notification, nil)
	}
}

func (di *CommentAPI) downloadAssetFromUrl(from string) error {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Parse the filename
	u, err := url.Parse(from)
	tokens := strings.Split(u.Path, "/")
	fileName := tokens[len(tokens)-1]

	response, err := client.Get(from)
	if err != nil {
		return errors.New(fmt.Sprint("Error while downloading", from, "-", err))
	}

	// Read all the bytes to the image
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.New(fmt.Sprint("Error while downloading", from, "-", err))
	}

	// Get the file type
	fileNameDots := strings.Split(fileName, ".")
	fileNameExt  := fileNameDots[len(fileNameDots)-1]
	fileMime := "." + mime.TypeByExtension(fileNameExt)
	path := "posts/" + fileName

	// Use s3 instance to save the file
	err = di.S3Bucket.Put(path, data, fileMime, s3.ACL("public-read"))

	if err != nil {
		panic(err)
	}

	// Close the request channel when needed
	response.Body.Close()

	return nil
}

func (di *CommentAPI) mentionUserComment(mentioned string, post model.Post, user_id bson.ObjectId) {

	// Get the comment author
	var user model.User
	var author model.User
	var notifications model.UserFirebaseNotifications

	// Get the author user
	err := di.DataService.Database.C("users").Find(bson.M{"_id": user_id}).One(&author)

	if err != nil {
		return
	}

	// Get the mentioned user
	username := mentioned[1:]
	err = di.DataService.Database.C("users").Find(bson.M{"username": username}).One(&user)

	if err == nil {

		// Gravatar url
		emailHash := gravatar.EmailHash(user.Email)
		image := gravatar.GetAvatarURL("http", emailHash, "http://spartangeek.com/images/default-avatar.png", 80)

		// Construct the notification message
		title := fmt.Sprintf("**%s** te mencionó en un comentario", author.UserName)
		message := post.Title

		// Process notification using firebase
		userPath := "users/" + user.Id.Hex() + "/notifications"
		userRef := di.Firebase.Child(userPath, nil, &notifications)

		userRef.Set("count", notifications.Count+1, nil)

		notification := &model.UserFirebaseNotification{
			UserId:    user_id,
			RelatedId: post.Id,
			RelatedExtra: post.Slug,
			Position:  post.Comments.Count,
			Title:     title,
			Text:      message,
			Related:   "mention",
			Seen:      false,
			Image:     image.String(),
			Created:   time.Now(),
			Updated:   time.Now(),
		}

		userRef.Child("list", nil, nil).Push(notification, nil)
	}
}