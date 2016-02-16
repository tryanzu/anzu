package handle

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/cosn/firebase"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/transmit"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
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
	Firebase      *firebase.Client             `inject:""`
	Gaming        *gaming.Module               `inject:""`
	ConfigService *config.Config               `inject:""`
	Transmit      *transmit.Sender             `inject:""`
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

		if err != nil || lim <= 0 {
			c.JSON(401, gin.H{"message": "Invalid request, check params.", "status": "error", "code": 901})
			return
		}

		limit = lim
	}

	search := make(bson.M)

	if f != "" && bson.IsObjectIdHex(f) {

		search["category"] = bson.ObjectIdHex(f)

		if signed_in {

			// Reset the counter for the user
			//di.resetUserCategoryCounter(f, bson.ObjectIdHex(user_id.(string)))
		}
	}

	if after != "" {

		t, err := time.Parse(time.RFC3339Nano, after)

		if err == nil {

			if signed_in {

				userPath := "users/" + user_id.(string)
				userRef := di.Firebase.Child(userPath, nil, nil)

				userRef.Set("pending", 0, nil)
			}

			search["created_at"] = bson.M{"$lt": t}
		}
	}

	if before != "" {

		t, err := time.Parse(time.RFC3339Nano, before)

		if err == nil {
			search["created_at"] = bson.M{"$gt": t}
		}
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
		search["deleted"] = bson.M{"$exists": false}

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
					Id:        postUser.Id,
					UserName:  postUser.UserName,
					FirstName: postUser.FirstName,
					LastName:  postUser.LastName,
					Step:      authorLevel,
					Email:     postUser.Email,
					Image:     postUser.Image,
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
	post, err := di.Feed.LightPost(post_id)

	if err != nil {

		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	data := post.Data()
	data.Content = html.UnescapeString(data.Content)

	c.JSON(200, data)
}

func (di PostAPI) PostsGetOne(c *gin.Context) {

	var legalSlug = regexp.MustCompile(`^([a-zA-Z0-9\-\.|/]+)$`)
	var err error

	// Get the database interface from the DI
	database := di.DataService.Database

	// Get the post using the slug
	id := c.Params.ByName("id")
	post_type := ""

	if bson.IsObjectIdHex(id) {
		post_type = "id"
	}

	if legalSlug.MatchString(id) && post_type == "" {
		post_type = "slug"
	}

	if post_type == "" {
		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	// Get the collection
	collection := database.C("posts")
	post := model.Post{}

	// Try to fetch the needed post by id
	if post_type == "id" {
		err = collection.FindId(bson.ObjectIdHex(id)).Select(bson.M{"comments.set": bson.M{"$slice": -10}}).One(&post)
	}

	if post_type == "slug" {
		err = collection.Find(bson.M{"slug": id}).Select(bson.M{"comments.set": bson.M{"$slice": -10}}).One(&post)
	}

	if err != nil {
		c.JSON(404, gin.H{"message": "Couldnt found post with that slug.", "status": "error"})
		return
	}

	// Get the users and stuff
	if post.Users != nil && len(post.Users) > 0 {

		var users []model.User

		// Get the users
		collection := database.C("users")

		err := collection.Find(bson.M{"_id": bson.M{"$in": post.Users}}).All(&users)

		if err != nil {
			panic(err)
		}

		usersMap := make(map[bson.ObjectId]interface{})

		var description string

		for _, user := range users {
			description = "Solo otro Spartan Geek más"

			if len(user.Description) > 0 {
				description = user.Description
			}

			usersMap[user.Id] = map[string]interface{}{
				"id":          user.Id.Hex(),
				"username":    user.UserName,
				"description": description,
				"image":       user.Image,
				"level":       user.Gaming.Level,
				"roles":       user.Roles,
			}

			if user.Id == post.UserId {
				// Set the author
				post.Author = user
			}
		}

		// Name of the set to get
		_, signed_in := c.Get("token")

		// Look for votes that has been already given
		var votes []model.Vote
		var likes []model.Vote
		var liked model.Vote

		if signed_in {

			user_id := c.MustGet("user_id")
			user_bson_id := bson.ObjectIdHex(user_id.(string))

			err = database.C("votes").Find(bson.M{"type": "component", "related_id": post.Id, "user_id": user_bson_id}).All(&votes)

			// Get the likes given by the current user
			_ = database.C("votes").Find(bson.M{"type": "comment", "related_id": post.Id, "user_id": user_bson_id}).All(&likes)

			err = database.C("votes").Find(bson.M{"type": "post", "related_id": post.Id, "user_id": user_bson_id}).One(&liked)

			if err == nil {

				post.Liked = liked.Value
			}

			// Increase user saw posts and its gamification in another thread
			go func(user_id bson.ObjectId, users []model.User) {

				var target model.User

				// Update the user saw posts
				_ = database.C("users").Update(bson.M{"_id": user_id}, bson.M{"$inc": bson.M{"stats.saw": 1}})
				player := false

				for _, user := range users {

					if user.Id == user_id {

						// The user is a player of the post so we dont have to get it from the database again
						player = true
						target = user
					}
				}

				if player == false {

					err = collection.Find(bson.M{"_id": user_id}).One(&target)

					if err != nil {
						panic(err)
					}
				}

				// Update user achievements (saw posts)
				//updateUserAchievement(target, "saw")

			}(user_bson_id, users)
		}

		if post.Solved == true {

			var bestComments model.CommentsPost

			err := database.C("posts").FindId(post.Id).Select(bson.M{"_id": 1, "comments.set": bson.M{"$elemMatch": bson.M{"chosen": true}}}).One(&bestComments)

			if err == nil && len(bestComments.Comments.Set) > 0 {
				post.Comments.Answer = bestComments.Comments.Set[0]
			}
		}

		// This will calculate the position based on the sliced array
		true_count := di.Feed.TrueCommentCount(post.Id)
		count := true_count - 10

		if count < 0 {
			count = 0
		}

		for index := range post.Comments.Set {

			comment := &post.Comments.Set[index]

			// Save the position over the comment
			post.Comments.Set[index].Position = count + index

			// Check if user liked that comment already
			for _, vote := range likes {

				if vote.NestedType == strconv.Itoa(index) {

					post.Comments.Set[index].Liked = vote.Value
				}
			}

			if _, okay := usersMap[comment.UserId]; okay {

				post.Comments.Set[index].User = usersMap[comment.UserId]
			}
		}

		// Remove deleted comments from the set
		comments := post.Comments.Set[:0]

		for _, c := range post.Comments.Set {

			if c.Deleted.IsZero() == true {

				comments = append(comments, c)
			}
		}

		post.Comments.Set = comments
		post.Comments.Total = true_count

		// Sort by created at
		sort.Sort(model.ByCommentCreatedAt(post.Comments.Set))

		// Get components information if components publication
		components := reflect.ValueOf(&post.Components).Elem()
		components_type := reflect.TypeOf(&post.Components).Elem()

		for i := 0; i < components.NumField(); i++ {

			f := components.Field(i)
			t := components_type.Field(i)

			if f.Type().String() == "model.Component" {

				component := f.Interface().(model.Component)

				for _, vote := range votes {

					if vote.NestedType == strings.ToLower(t.Name) {

						if vote.Value == 1 {

							component.Voted = "up"

						} else if vote.Value == -1 {

							component.Voted = "down"
						}
					}
				}

				if component.Elections == true {

					for option_index, option := range component.Options {

						if _, okay := usersMap[option.UserId]; okay {

							component.Options[option_index].User = usersMap[option.UserId]
						}
					}

					// Sort by created at
					sort.Sort(model.ByElectionsCreatedAt(component.Options))
				}

				f.Set(reflect.ValueOf(component))
			}
		}
	}

	// Save the activity
	signed_id, signed_in := c.Get("user_id")
	user_id := ""

	if signed_in {

		user_id = signed_id.(string)
	}

	go func(post model.Post, user_id string, signed_in bool) {

		defer di.Errors.Recover()

		post_module, _ := di.Feed.Post(post)

		if signed_in {

			by := bson.ObjectIdHex(user_id)

			post_module.Viewed(by)
		}

		post_module.UpdateRate()

		// Trigger gamification events (if needed)
		di.Gaming.Post(post_module).Review()

	}(post, user_id, signed_in)

	c.JSON(200, post)
}

func (di PostAPI) PostCreate(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	// Check for user token
	user_id, _ := c.Get("user_id")
	bson_id := bson.ObjectIdHex(user_id.(string))

	var post model.PostForm

	// Get the form otherwise tell it has been an error
	if c.BindWith(&post, binding.JSON) == nil {

		post_category := post.Category

		if bson.IsObjectIdHex(post_category) == false {

			c.JSON(400, gin.H{"status": "error", "message": "Invalid category id"})
			return
		}

		var category model.Category

		err := database.C("categories").Find(bson.M{"parent": bson.M{"$exists": true}, "_id": bson.ObjectIdHex(post_category)}).One(&category)

		if err != nil {

			c.JSON(400, gin.H{"status": "error", "message": "Invalid category"})
			return
		}

		user := di.Acl.User(bson_id)

		if user.CanWrite(category) == false {

			c.JSON(400, gin.H{"status": "error", "message": "Not enough permissions."})
			return
		}

		if post.Pinned == true && user.Can("pin-board-posts") == false {

			c.JSON(400, gin.H{"status": "error", "message": "Not enough permissions to pin."})
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

		content := html.EscapeString(post.Content)
		urls, _ := regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
		post_id := bson.NewObjectId()

		var assets []string

		assets = urls.FindAllString(content, -1)

		// Empty participants list - only author included
		users := []bson.ObjectId{bson_id}

		switch post.Kind {
		case "recommendations":

			components := post.Components

			if len(components) > 0 {

				budget, bo := components["budget"]
				budget_type, bto := components["budget_type"]
				budget_currency, bco := components["budget_currency"]
				budget_flexibility, bfo := components["budget_flexibility"]
				software, so := components["software"]

				// Clean up components for further speed checking
				delete(components, "budget")
				delete(components, "budget_type")
				delete(components, "budget_currency")
				delete(components, "budget_flexibility")
				delete(components, "software")

				if !bo || !bto || !bco || !bfo || !so {

					// Some important information is missing for this kind of post
					c.JSON(400, gin.H{"status": "error", "message": "Couldnt create post, missing information...", "code": 4001})
					return
				}

				post_name := "PC '" + post.Name
				if budget.(string) != "0" {
					post_name += "' con presupuesto de $" + budget.(string) + " " + budget_currency.(string)
				} else {
					post_name += "'"
				}

				slug := helpers.StrSlug(post_name)
				slug_exists, _ := database.C("posts").Find(bson.M{"slug": slug}).Count()

				if slug_exists > 0 {

					slug = helpers.StrSlugRandom(post_name)
				}

				publish := &model.Post{
					Id:         post_id,
					Title:      post_name,
					Content:    content,
					Type:       "recommendations",
					Slug:       slug,
					Comments:   comments,
					UserId:     bson_id,
					Users:      users,
					Categories: []string{"recommendations"},
					Category:   bson.ObjectIdHex(post_category),
					Votes:      votes,
					IsQuestion: post.IsQuestion,
					Pinned:     post.Pinned,
					Created:    time.Now(),
					Updated:    time.Now(),
				}

				publish_components := model.Components{
					Budget:            budget.(string),
					BudgetType:        budget_type.(string),
					BudgetCurrency:    budget_currency.(string),
					BudgetFlexibility: budget_flexibility.(string),
					Software:          software.(string),
				}

				for component, value := range components {

					component_elements := value.(map[string]interface{})
					bindable := reflect.ValueOf(&publish_components).Elem()

					for i := 0; i < bindable.NumField(); i++ {

						t := bindable.Type().Field(i)
						json_tag := t.Tag
						name := json_tag.Get("json")
						status := "owned"

						if component_elements["owned"].(bool) == false {
							status = "needed"
						}

						if name == component || name == component+",omitempty" {

							c := model.Component{
								Elections: component_elements["poll"].(bool),
								Status:    status,
								Votes:     votes,
								Content:   component_elements["value"].(string),
							}

							// Set the component with the component we've build above
							bindable.Field(i).Set(reflect.ValueOf(c))
						}
					}
				}

				// Now bind the components to the post
				publish.Components = publish_components

				err := database.C("posts").Insert(publish)

				if err != nil {
					panic(err)
				}

				for _, asset := range assets {

					// Download the asset on other routine in order to non block the API request
					go di.downloadAssetFromUrl(asset, publish.Id)
				}

				// Add the gamification contribution
				go di.Gaming.Get(bson_id).Did("publish")

				// Add a counter for the category
				di.addUserCategoryCounter("recommendations")

				// Sync everyone's feed
				go di.syncUsersFeed(publish)

				go func(id bson.ObjectId, module *feed.FeedModule) {

					post, err := module.Post(id)

					if err != nil {
						panic(err)
					}

					// Index the brand new post
					post.Index()

				}(post_id, di.Feed)

				// Finished creating the post
				c.JSON(200, gin.H{"status": "okay", "code": 200, "post": gin.H{"id": post_id, "slug": slug}})
				return
			}

		case "category-post":

			title := post.Name

			if len([]rune(title)) > 72 {
				title = helpers.Truncate(title, 72) + "..."
			}

			slug := helpers.StrSlug(title)
			slug_exists, _ := database.C("posts").Find(bson.M{"slug": slug}).Count()

			if slug_exists > 0 {

				slug = helpers.StrSlugRandom(title)
			}

			publish := &model.Post{
				Id:         post_id,
				Title:      title,
				Content:    content,
				Type:       "category-post",
				Slug:       slug,
				Comments:   comments,
				UserId:     bson_id,
				Users:      users,
				Categories: []string{post.Tag},
				Category:   bson.ObjectIdHex(post_category),
				Votes:      votes,
				IsQuestion: post.IsQuestion,
				Pinned:     post.Pinned,
				Created:    time.Now(),
				Updated:    time.Now(),
			}

			err := database.C("posts").Insert(publish)

			if err != nil {
				panic(err)
			}

			for _, asset := range assets {

				// Download the asset on other routine in order to non block the API request
				go di.downloadAssetFromUrl(asset, publish.Id)
			}

			// Add the gamification contribution
			go di.Gaming.Get(bson_id).Did("publish")

			// Add a counter for the category
			di.addUserCategoryCounter(post.Tag)

			// Sync everyone's feed
			go di.syncUsersFeed(publish)

			go func(id bson.ObjectId, module *feed.FeedModule) {

				post, err := module.Post(id)

				if err != nil {
					panic(err)
				}

				// Index the brand new post
				post.Index()

			}(post_id, di.Feed)

			// Finished creating the post
			c.JSON(200, gin.H{"status": "okay", "code": 200, "post": gin.H{"id": post_id, "slug": slug}})
			return
		}
	}

	c.JSON(400, gin.H{"status": "error", "message": "Couldnt create post, missing information...", "code": 205})
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

func (di PostAPI) PostUpdate(c *gin.Context) {

	var post model.Post
	var postForm model.PostForm

	// Get the database interface from the DI
	database := di.DataService.Database

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

		user := di.Acl.User(user_bson_id)

		if user.CanUpdatePost(post) == false {

			c.JSON(400, gin.H{"message": "Can't update post. Insufficient permissions", "status": "error"})
			return
		}

		if user.CanWrite(category) == false {

			c.JSON(400, gin.H{"status": "error", "message": "Not enough permissions to write this category."})
			return
		}

		if postForm.Pinned == true && postForm.Pinned != post.Pinned && user.Can("pin-board-posts") == false {

			c.JSON(400, gin.H{"status": "error", "message": "Not enough permissions to pin."})
			return
		}

		slug := helpers.StrSlug(postForm.Name)
		slug_exists, _ := database.C("posts").Find(bson.M{"slug": slug}).Count()

		if slug_exists > 0 {

			slug = helpers.StrSlugRandom(postForm.Name)
		}

		content := html.EscapeString(postForm.Content)
		urls, _ := regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)

		var assets []string
		assets = urls.FindAllString(content, -1)

		update_directive := bson.M{"$set": bson.M{"content": content, "slug": slug, "title": postForm.Name, "category": bson.ObjectIdHex(post_category), "updated_at": time.Now()}}

		if postForm.Pinned == true {

			// Update the set directive by creating a copy of it and using type assertion
			set_directive := update_directive["$set"].(bson.M)
			set_directive["pinned"] = postForm.Pinned
			update_directive["$set"] = set_directive

			go func(carrier *transmit.Sender, id bson.ObjectId) {

				carrierParams := map[string]interface{}{
					"fire": "pinned",
					"id": id.Hex(),
				} 

				carrier.Emit("feed", "action", carrierParams)

				carrierParams = map[string]interface{}{
					"fire": "updated",
				} 

				carrier.Emit("post", id.Hex(), carrierParams)

			}(di.Transmit, post.Id)

		} else {

			update_directive["$unset"] = bson.M{"pinned": ""}

			go func(carrier *transmit.Sender, id bson.ObjectId) {

				carrierParams := map[string]interface{}{
					"fire": "unpinned",
					"id": id.Hex(),
				} 

				carrier.Emit("feed", "action", carrierParams)

			}(di.Transmit, post.Id)
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

		}(post.Id, di.Feed)

		if err != nil {
			panic(err)
		}

		for _, asset := range assets {

			// Download the asset on other routine in order to non block the API request
			go di.downloadAssetFromUrl(asset, post.Id)
		}

		c.JSON(200, gin.H{"status": "okay"})
		return
	}

	c.JSON(400, gin.H{"status": "error", "message": "Couldnt update post, missing information..."})
}

func (di PostAPI) PostDelete(c *gin.Context) {

	var post model.Post

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
	err := database.C("posts").FindId(bson_id).One(&post)

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

	go func(carrier *transmit.Sender, id bson.ObjectId) {

		carrierParams := map[string]interface{}{
			"fire": "delete-post",
			"id": id.Hex(),
		} 

		carrier.Emit("feed", "action", carrierParams)

	}(di.Transmit, post.Id)

	c.JSON(200, gin.H{"status": "okay"})
}

func (di PostAPI) syncUsersFeed(post *model.Post) {

	var users map[string]model.UserFirebase

	redis := di.CacheService
	carrier := di.Transmit

	// Recover from any panic even inside this goroutine
	defer di.Errors.Recover()

	carrierParams := map[string]interface{}{
		"fire": "new-post",
		"category": post.Category.Hex(),
	} 

	carrier.Emit("feed", "action", carrierParams)

	// Search the online users
	onlineParams := map[string]string{
		"orderBy": "\"online\"",
		"startAt": "1",
	}
	_ = di.Firebase.Child("users", onlineParams, &users)

	// Information about the post
	category := post.Category.Hex()

	for user_id, user := range users {

		// Must be either seeing that category or own general feed
		if user.Viewing != category && user.Viewing != "all" {
			continue
		}

		if user.Viewing == "all" {

			subscribed, err := redis.SIsMember("user:categories:"+user_id, category)

			// User is actually not subscribed or and error just happened
			if subscribed == false || err != nil {

				// Temp stuff - check if user has no single subscription already
				user_categories_list, err := redis.SMembers("user:categories:" + user_id)

				if err != nil || len(user_categories_list) > 0 {
					continue
				}
			}
		}

		// Add a pending counter
		userPath := "users/" + user_id
		userRef := di.Firebase.Child(userPath, nil, nil)

		userRef.Set("pending", user.Pending+1, nil)
	}
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
