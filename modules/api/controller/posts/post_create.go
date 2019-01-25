package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/helpers"
	"gopkg.in/mgo.v2/bson"

	"html"
	"net/http"
	"regexp"
	"time"
)

func (this API) Create(c *gin.Context) {
	var form model.PostForm
	database := deps.Container.Mgo()

	// Check for user token
	user_id, _ := c.Get("user_id")
	uid := bson.ObjectIdHex(user_id.(string))

	// Get the form otherwise tell it has been an error
	if c.BindWith(&form, binding.JSON) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Couldnt create post, missing information...", "code": 205})
		return
	}

	if bson.IsObjectIdHex(form.Category) == false {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid category id"})
		return
	}

	if len(form.Content) > 25000 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "too much content. excedeed limit of 25,000 characters"})
		return
	}

	var category model.Category
	err := database.C("categories").Find(bson.M{
		"parent": bson.M{"$exists": true},
		"_id":    bson.ObjectIdHex(form.Category),
	}).One(&category)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid category"})
		return
	}

	usr := this.Acl.User(uid)
	if usr.CanWrite(category) == false || usr.HasValidated() == false {
		c.JSON(403, gin.H{"status": "error", "message": "Not enough permissions."})
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
	urls, _ := regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
	post_id := bson.NewObjectId()

	var assets []string

	assets = urls.FindAllString(content, -1)

	// Empty participants list - only author included
	users := []bson.ObjectId{uid}
	title := form.Title
	if len([]rune(title)) > 72 {
		title = helpers.Truncate(title, 72) + "..."
	}

	slug := helpers.StrSlug(title)
	if c, _ := database.C("posts").Find(bson.M{"slug": slug}).Count(); c > 0 {
		slug = helpers.StrSlugRandom(title)
	}

	publish := &model.Post{
		Id:         post_id,
		Title:      title,
		Content:    content,
		Type:       "category-post",
		Slug:       slug,
		Comments:   comments,
		UserId:     uid,
		Users:      users,
		Category:   bson.ObjectIdHex(form.Category),
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

	err = database.C("posts").Insert(publish)
	if err != nil {
		panic(err)
	}

	// Notify events pool immediately after performing save.
	events.In <- events.PostNew(publish.Id)

	for _, asset := range assets {

		// Non blocking image download
		go this.savePostImages(asset, publish.Id)
	}

	// Finished creating the post
	c.JSON(200, gin.H{"status": "okay", "code": 200, "post": gin.H{"id": post_id, "slug": slug}})
}
