package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/core/content"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/feed"
	"github.com/tryanzu/core/modules/helpers"
	"gopkg.in/mgo.v2/bson"

	"html"
	"net/http"
	"time"
)

func (this API) Update(c *gin.Context) {
	var form model.PostForm

	// Get the post using the id
	id := c.Params.ByName("id")
	if bson.IsObjectIdHex(id) == false {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request, no valid params.", "status": "error"})
		return
	}

	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "err": err})
		return
	}

	if len(form.Content) > 25000 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "too much content. excedeed limit of 25,000 chars"})
		return
	}

	// Get the post using the slug
	uid := c.MustGet("userID").(bson.ObjectId)
	bid := bson.ObjectIdHex(id)
	post, err := this.Feed.Post(bid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Couldnt find the post", "status": "error"})
		return
	}

	if bson.IsObjectIdHex(form.Category) == false {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid category id"})
		return
	}

	var category model.Category
	err = deps.Container.Mgo().C("categories").Find(bson.M{"parent": bson.M{"$exists": true}, "_id": bson.ObjectIdHex(form.Category)}).One(&category)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid category"})
		return
	}

	user := this.Acl.User(uid)
	if user.CanUpdatePost(post) == false {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Can't update post. Insufficient permissions", "status": "error"})
		return
	}

	if post.Category != category.Id && user.CanWrite(category.Permissions.Write) == false {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Not enough permissions to write this category."})
		return
	}

	if form.Lock == true && form.Lock != post.Lock && user.CanLockPost(post) == false {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Not enough permissions to block comments in this post"})
		return
	}

	if form.Pinned == true && form.Pinned != post.Pinned && user.Can("pin-board-posts") == false {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Not enough permissions to pin."})
		return
	}

	slug := post.Slug
	if form.Title != post.Title {
		slug := helpers.StrSlug(form.Title)
		if duplicated, _ := deps.Container.Mgo().C("posts").Find(bson.M{"slug": slug}).Count(); duplicated > 0 {
			slug = helpers.StrSlugRandom(form.Title)
		}

		events.In <- events.RawEmit("feed", "action", map[string]interface{}{
			"fire":  "changed-title",
			"id":    post.Id.Hex(),
			"title": form.Title,
			"slug":  slug,
		})
	}

	post.Content = html.EscapeString(form.Content)
	// Pre-process comment content.
	processed, err := content.Preprocess(deps.Container, post)
	if err != nil {
		return
	}

	post = processed.(*feed.Post)
	query := bson.M{
		"$set": bson.M{
			"content":    post.Content,
			"slug":       slug,
			"title":      form.Title,
			"category":   bson.ObjectIdHex(form.Category),
			"updated_at": time.Now(),
		},
	}
	unset := bson.M{}
	if form.Pinned == true {
		// Update the set directive by creating a copy of it and using type assertion
		set := query["$set"].(bson.M)
		set["pinned"] = form.Pinned
		query["$set"] = set
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

	if form.Lock == true {
		set := query["$set"].(bson.M)
		set["lock"] = form.Lock
		query["$set"] = set
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
		query["$unset"] = unset
	}

	if form.IsQuestion != post.IsQuestion {
		set_directive := query["$set"].(bson.M)
		set_directive["is_question"] = form.IsQuestion
		query["$set"] = set_directive
	}

	err = deps.Container.Mgo().C("posts").Update(bson.M{"_id": post.Id}, query)
	if err != nil {
		panic(err)
	}

	events.In <- events.RawEmit("post", post.Id.Hex(), map[string]interface{}{
		"fire": "updated",
	})

	c.JSON(http.StatusOK, gin.H{"status": "okay", "id": post.Id.Hex(), "slug": post.Slug})
}
