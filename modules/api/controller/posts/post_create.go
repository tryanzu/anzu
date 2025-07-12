package posts

import (
	"context"
	"html"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	assetURL, _ = regexp.Compile(`^http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
)

func (this API) Create(c *gin.Context) {
	var form model.PostForm

	// Check for user token
	uid, err := primitive.ObjectIDFromHex(c.MustGet("user_id").(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid user ID"})
		return
	}

	// Get the form otherwise tell it has been an error
	if c.BindWith(&form, binding.JSON) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Couldnt create post, missing information...", "code": 205})
		return
	}

	if _, err := primitive.ObjectIDFromHex(form.Category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid category id"})
		return
	}

	if len(form.Content) > 25000 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "too much content. excedeed limit of 25,000 characters"})
		return
	}

	ctx := context.Background()
	var category model.Category
	categoryID, _ := primitive.ObjectIDFromHex(form.Category)
	database := deps.Container.Mgo()
	categoriesCollection := database.Collection("categories")
	filter := bson.M{
		"parent": bson.M{"$exists": true},
		"_id":    categoryID,
	}
	err = categoriesCollection.FindOne(ctx, filter).Decode(&category)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid category"})
		return
	}

	usr := this.Acl.User(uid)
	if usr.CanWrite(category.Permissions.Write) == false || usr.HasValidated() == false {
		c.JSON(403, gin.H{"status": "error", "message": "Not enough permissions to post in this category."})
		return
	}

	if form.Pinned == true && usr.Can("pin-board-posts") == false {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Not enough permissions to pin."})
		return
	}

	if form.Lock == true && usr.Can("block-own-post-comments") == false {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Not enough permissions to lock."})
		return
	}

	comments := model.Comments{
		Count: 0,
		Set:   make([]model.Comment, 0),
	}

	votes := model.Votes{
		Up:     0,
		Down:   0,
		Rating: 0,
	}

	content := html.EscapeString(form.Content)

	var assets []string
	assets = assetURL.FindAllString(content, -1)

	// Empty participants list - only author included
	users := []primitive.ObjectID{uid}
	title := form.Title
	if len([]rune(title)) > 72 {
		title = helpers.Truncate(title, 72) + "..."
	}

	slug := helpers.StrSlug(title)
	postsCollection := database.Collection("posts")
	if c, _ := postsCollection.CountDocuments(ctx, bson.M{"slug": slug}); c > 0 {
		slug = helpers.StrSlugRandom(title)
	}

	publish := model.Post{
		Id:         primitive.NewObjectID(),
		Title:      title,
		Content:    content,
		Type:       "category-post",
		Slug:       slug,
		Comments:   comments,
		UserId:     uid,
		Users:      users,
		Category:   categoryID,
		Votes:      votes,
		IsQuestion: form.IsQuestion,
		Pinned:     form.Pinned,
		Lock:       form.Lock,
		Created:    time.Now(),
		Updated:    time.Now(),
	}

	u, err := user.FindId(deps.Container, uid)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	if user.CanBeTrusted(u) == false {
		publish.Deleted = time.Now()
	}

	_, err = postsCollection.InsertOne(ctx, &publish)
	if err != nil {
		// Check for MongoDB write errors
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"status": "error", "message": "Post already exists", "code": 409})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to create post", "code": 500})
		return
	}

	// Notify events pool immediately after performing save.
	events.In <- events.PostNew(publish.Id)

	for _, asset := range assets {

		// Non blocking image download
		go this.savePostImages(asset, publish.Id)
	}

	// Finished creating the post
	c.JSON(200, gin.H{"status": "okay", "code": 200, "post": gin.H{"id": publish.Id, "slug": slug}})
}
