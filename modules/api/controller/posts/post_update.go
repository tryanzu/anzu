package posts


import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/transmit"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/mgo.v2/bson"

	"html"
	"regexp"
	"time"
)

func (this API) Update(c *gin.Context) {

	var post model.Post
	var postForm model.PostForm

	// Get the database interface from the DI
	database := this.Mongo.Database

	// Get the post using the id
	id := c.Params.ByName("id")

	if bson.IsObjectIdHex(id) == false {

		c.JSON(400, gin.H{"message": "Invalid request, no valid params.", "status": "error"})
		return
	}

	// Get the form otherwise tell it has been an error
	if c.BindWith(&postForm, binding.JSON) == nil {

		// Get the post using the slug
		user_id := c.MustGet("user_id")
		user_bson_id := bson.ObjectIdHex(user_id.(string))
		bson_id := bson.ObjectIdHex(id)
		err := database.C("posts").FindId(bson_id).One(&post)

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

		if user.CanWrite(category) == false {

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

		if postForm.Name != post.Title {

			slug := helpers.StrSlug(postForm.Name)
			slug_exists, _ := database.C("posts").Find(bson.M{"slug": slug}).Count()

			if slug_exists > 0 {
				slug = helpers.StrSlugRandom(postForm.Name)
			}
		}

		content := html.EscapeString(postForm.Content)
		urls, _ := regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)

		var assets []string
		assets = urls.FindAllString(content, -1)

		update_directive := bson.M{"$set": bson.M{"content": content, "slug": slug, "title": postForm.Name, "category": bson.ObjectIdHex(post_category), "updated_at": time.Now()}}
		unset := bson.M{}

		if postForm.Pinned == true {

			// Update the set directive by creating a copy of it and using type assertion
			set_directive := update_directive["$set"].(bson.M)
			set_directive["pinned"] = postForm.Pinned
			update_directive["$set"] = set_directive

			if post.Pinned == false {			
				go func(carrier *transmit.Sender, id bson.ObjectId) {

					carrierParams := map[string]interface{}{
						"fire": "pinned",
						"id": id.Hex(),
					} 

					carrier.Emit("feed", "action", carrierParams)

				}(this.Transmit, post.Id)
			}

		} else {

			unset["pinned"] = ""

			if post.Pinned == true {
				go func(carrier *transmit.Sender, id bson.ObjectId) {

					carrierParams := map[string]interface{}{
						"fire": "unpinned",
						"id": id.Hex(),
					} 

					carrier.Emit("feed", "action", carrierParams)

				}(this.Transmit, post.Id)
			}
		}

		if postForm.Lock == true {

			set_directive := update_directive["$set"].(bson.M)
			set_directive["lock"] = postForm.Lock
			update_directive["$set"] = set_directive

			if post.Lock == false {			
				go func(carrier *transmit.Sender, id bson.ObjectId) {

					carrierParams := map[string]interface{}{
						"fire": "locked",
					} 

					carrier.Emit("post", id.Hex(), carrierParams)

				}(this.Transmit, post.Id)
			}

		} else {

			unset["lock"] = ""

			if post.Lock == true {
				go func(carrier *transmit.Sender, id bson.ObjectId) {

					carrierParams := map[string]interface{}{
						"fire": "unlocked",
					} 

					carrier.Emit("post", id.Hex(), carrierParams)

				}(this.Transmit, post.Id)
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

		go func(id bson.ObjectId, module *feed.FeedModule) {

			post, err := module.Post(id)

			if err != nil {
				panic(err)
			}

			// Index the brand new post
			post.Index()

		}(post.Id, this.Feed)

		if err != nil {
			panic(err)
		}

		for _, asset := range assets {

			// Download the asset on other routine in order to non block the API request
			go this.savePostImages(asset, post.Id)
		}

		go func(carrier *transmit.Sender, id bson.ObjectId) {

			carrierParams := map[string]interface{}{
				"fire": "updated",
			} 

			carrier.Emit("post", id.Hex(), carrierParams)

		}(this.Transmit, post.Id)

		c.JSON(200, gin.H{"status": "okay", "id": post.Id.Hex(), "slug": post.Slug})
		return
	}

	c.JSON(400, gin.H{"status": "error", "message": "Couldnt update post, missing information..."})
}