package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/helpers"
	"gopkg.in/mgo.v2/bson"

	"html"
	"regexp"
	"time"
)

func (this API) Update(c *gin.Context) {
	var postForm model.PostForm

	// Get the database interface from the DI
	database := deps.Container.Mgo()

	// Get the post using the id
	id := c.Params.ByName("id")

	if bson.IsObjectIdHex(id) == false {
		c.JSON(400, gin.H{"message": "Invalid request, no valid params.", "status": "error"})
		return
	}

	if err := c.BindJSON(&postForm); err != nil {
		c.JSON(400, gin.H{"status": "error", "err": err})
		return
	}

	// Get the post using the slug
	user_id := c.MustGet("user_id")
	user_bson_id := bson.ObjectIdHex(user_id.(string))
	bson_id := bson.ObjectIdHex(id)
	post, err := this.Feed.Post(bson_id)

	if err != nil {
		c.JSON(404, gin.H{"message": "Couldnt find the post", "status": "error"})
		return
	}

	post_category := postForm.Category

	if bson.IsObjectIdHex(post_category) == false {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid category id"})
		return
	}

	var category model.Category
	err = database.C("categories").Find(bson.M{"parent": bson.M{"$exists": true}, "_id": bson.ObjectIdHex(post_category)}).One(&category)

	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid category"})
		return
	}

	user := this.Acl.User(user_bson_id)
	if user.CanUpdatePost(post) == false {
		c.JSON(400, gin.H{"message": "Can't update post. Insufficient permissions", "status": "error"})
		return
	}

	if post.Category != category.Id && user.CanWrite(category) == false {
		c.JSON(400, gin.H{"status": "error", "message": "Not enough permissions to write this category."})
		return
	}

	if postForm.Lock == true && postForm.Lock != post.Lock && user.CanLockPost(post) == false {
		c.JSON(400, gin.H{"status": "error", "message": "Not enough permissions to lock."})
		return
	}

	if postForm.Pinned == true && postForm.Pinned != post.Pinned && user.Can("pin-board-posts") == false {
		c.JSON(400, gin.H{"status": "error", "message": "Not enough permissions to pin."})
		return
	}

	slug := post.Slug

	if postForm.Title != post.Title {
		slug := helpers.StrSlug(postForm.Title)
		slug_exists, _ := database.C("posts").Find(bson.M{"slug": slug}).Count()

		if slug_exists > 0 {
			slug = helpers.StrSlugRandom(postForm.Title)
		}

		events.In <- events.RawEmit("feed", "action", map[string]interface{}{
			"fire":  "changed-title",
			"id":    post.Id.Hex(),
			"title": postForm.Title,
			"slug":  slug,
		})
	}

	content := html.EscapeString(postForm.Content)
	urls, _ := regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)

	var assets []string
	assets = urls.FindAllString(content, -1)

	update_directive := bson.M{"$set": bson.M{"content": content, "slug": slug, "title": postForm.Title, "category": bson.ObjectIdHex(post_category), "updated_at": time.Now()}}
	unset := bson.M{}

	if postForm.Pinned == true {

		// Update the set directive by creating a copy of it and using type assertion
		set := update_directive["$set"].(bson.M)
		set["pinned"] = postForm.Pinned
		update_directive["$set"] = set

		if post.Pinned == false {
			events.In <- events.RawEmit("feed", "action", map[string]interface{}{
				"fire": "pinned",
				"id":   post.Id.Hex(),
			})
		}

	} else {

		unset["pinned"] = ""

		if post.Pinned == true {
			events.In <- events.RawEmit("feed", "action", map[string]interface{}{
				"fire": "unpinned",
				"id":   post.Id.Hex(),
			})
		}
	}

	if postForm.Lock == true {
		set := update_directive["$set"].(bson.M)
		set["lock"] = postForm.Lock
		update_directive["$set"] = set

		if post.Lock == false {
			events.In <- events.RawEmit("post", post.Id.Hex(), map[string]interface{}{
				"fire": "locked",
			})
		}

	} else {

		unset["lock"] = ""

		if post.Lock == true {
			events.In <- events.RawEmit("post", post.Id.Hex(), map[string]interface{}{
				"fire": "unlocked",
			})
		}
	}

	if len(unset) > 0 {
		update_directive["$unset"] = unset
	}

	if postForm.IsQuestion != post.IsQuestion {

		set_directive := update_directive["$set"].(bson.M)
		set_directive["is_question"] = postForm.IsQuestion

		update_directive["$set"] = set_directive
	}

	err = database.C("posts").Update(bson.M{"_id": post.Id}, update_directive)

	if err != nil {
		panic(err)
	}

	for _, asset := range assets {

		// Download the asset on other routine in order to non block the API request
		go this.savePostImages(asset, post.Id)
	}

	events.In <- events.RawEmit("post", post.Id.Hex(), map[string]interface{}{
		"fire": "updated",
	})

	c.JSON(200, gin.H{"status": "okay", "id": post.Id.Hex(), "slug": post.Slug})
}
