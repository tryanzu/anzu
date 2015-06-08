package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/mrvdot/golang-utils"
	"github.com/kennygrant/sanitize"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/fernandez14/spartangeek-blacker/model"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type PostAPI struct {
	DataService *mongo.Service `inject:""`
}

/**
 * Avaliable components
 *
 * List of valid components along a recommendation post
 */
var avaliable_components = []string{
	"cpu", "motherboard", "ram", "cabinet", "screen", "storage", "cooler", "power", "videocard",
}

func (di *PostAPI) FeedGet(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	var feed []model.Post
	offset := 0
	limit := 10

	qs := c.Request.URL.Query()

	o := qs.Get("offset")
	l := qs.Get("limit")
	f := qs.Get("category")

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
	
	if f != "" && f != "recent" {
		
		search["categories"] = f
		
		// Check for user token
		user_token, signed_in := c.Get("user_id")
	
		if signed_in {
	
			// Reset the counter for the user			
			di.resetUserCategoryCounter(f, bson.ObjectIdHex(user_token.(string)))
		}
	}

	// Prepare the database to fetch the feed
	posts_collection := database.C("posts")
	get_feed := posts_collection.Find(search).Sort("-created_at").Limit(limit).Skip(offset)

	// Get the results from the feed algo
	err := get_feed.All(&feed)

	if err != nil {
		panic(err)
	}

	var authors []bson.ObjectId
	var users []model.User

	for _, post := range feed {

		authors = append(authors, post.UserId)
	}

	// Get the users needed by the feed
	err = database.C("users").Find(bson.M{"_id": bson.M{"$in": authors}}).All(&users)

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
				}
			}
		}

		c.JSON(200, gin.H{"feed": feed, "offset": offset, "limit": limit})
	} else {

		c.JSON(200, gin.H{"feed": []string{}, "offset": offset, "limit": limit})
	}
}

func (di *PostAPI) PostsGet(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	var recommendations []model.Post
	var published []model.Post

	// Get recommendation posts
	posts_collection := database.C("posts")
	get_recommendations := posts_collection.Find(bson.M{"type": "recommendations"}).Sort("-created_at").Limit(6)

	// Try to fetch the posts
	err := get_recommendations.All(&recommendations)

	if err != nil {
		panic(err)
	}

	get_published := posts_collection.Find(bson.M{"type": bson.M{"$ne": "recommendations"}}).Sort("-created_at").Limit(6)

	// Try to fetch the posts
	err = get_published.All(&published)

	if err != nil {

		panic(err)
	}

	var authors []bson.ObjectId

	for _, post := range recommendations {

		authors = append(authors, post.UserId)
	}

	for _, post := range published {

		authors = append(authors, post.UserId)
	}

	var users []model.User

	// Get the users
	collection := database.C("users")

	err = collection.Find(bson.M{"_id": bson.M{"$in": authors}}).All(&users)

	if err != nil {
		panic(err)
	}

	usersMap := make(map[bson.ObjectId]model.User)

	for _, user := range users {

		usersMap[user.Id] = user
	}

	for index := range recommendations {

		post := &recommendations[index]

		if _, okay := usersMap[post.UserId]; okay {

			postUser := usersMap[post.UserId]

			post.Author = model.User{
				Id:        postUser.Id,
				UserName:  postUser.UserName,
				FirstName: postUser.FirstName,
				LastName:  postUser.LastName,
			}
		}
	}

	for index := range published {

		post := &published[index]

		if _, okay := usersMap[post.UserId]; okay {

			postUser := usersMap[post.UserId]

			post.Author = model.User{
				Id:        postUser.Id,
				UserName:  postUser.UserName,
				FirstName: postUser.FirstName,
				LastName:  postUser.LastName,
			}
		}
	}

	c.JSON(200, gin.H{"recommendations": recommendations, "last": published})
}

func (di *PostAPI) PostsGetOne(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	// Get the post using the slug
	id := c.Params.ByName("id")

	if bson.IsObjectIdHex(id) == false {

		c.JSON(400, gin.H{"error": "Invalid request, id not valid.", "status": 400})

		return
	}

	// Get the collection
	collection := database.C("posts")

	post := model.Post{}

	// Try to fetch the needed post by id
	err := collection.FindId(bson.ObjectIdHex(id)).One(&post)

	if err != nil {

		c.JSON(404, gin.H{"error": "Couldnt found post with that slug.", "status": 203})

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

			if user.Id == post.UserId {

				// Set the author
				post.Author = user
			}

			description = "Solo otro mas Spartan Geek"

			if user_description, has_description := user.Profile["bio"]; has_description {

				description = user_description.(string)
			} 

			usersMap[user.Id] = map[string]string{
				"id":    user.Id.Hex(),
				"username":  user.UserName,
				"description": description,
				"email": user.Email,
			}
		}

		// Get the query parameters
		qs := c.Request.URL.Query()

		// Name of the set to get
		token := qs.Get("token")

		// Look for votes that has been already given
		var votes []model.Vote
		var likes []model.Vote

		if token != "" {

			// Get user by token
			user_token := model.UserToken{}

			// Try to fetch the user using token header though
			err = database.C("tokens").Find(bson.M{"token": token}).One(&user_token)

			if err == nil {

				err = database.C("votes").Find(bson.M{"type": "component", "related_id": post.Id, "user_id": user_token.UserId}).All(&votes)

				// Get the likes given by the current user
				_ = database.C("votes").Find(bson.M{"type": "comment", "related_id": post.Id, "user_id": user_token.UserId}).All(&likes)
			}

			if user_token.UserId != post.UserId {

				// Check if following
				following := model.UserFollowing{}

				err = database.C("followers").Find(bson.M{"follower": user_token.UserId, "following": post.UserId}).One(&following)

				// The user is following the author so tell the post struct
				if err == nil {

					post.Following = true
				}
			}

			// Increase user saw posts and its gamification in another thread
			go func(token model.UserToken, users []model.User) {

				var target model.User

				// Update the user saw posts
				_ = database.C("users").Update(bson.M{"_id": token.UserId}, bson.M{"$inc": bson.M{"stats.saw": 1}})
				player := false

				for _, user := range users {

					if user.Id == token.UserId {

						// The user is a player of the post so we dont have to get it from the database again
						player = true
						target = user
					}
				}

				if player == false {

					err = collection.Find(bson.M{"_id": token.UserId}).One(&target)

					if err != nil {
						panic(err)
					}
				}

				// Update user achievements (saw posts)
				//updateUserAchievement(target, "saw")

			}(user_token, users)
		}

		for index := range post.Comments.Set {

			comment := &post.Comments.Set[index]

			// Save the position over the comment
			post.Comments.Set[index].Position = index

			// Check if user liked that comment already
			for _, vote := range likes {

				if vote.NestedType == strconv.Itoa(index) {

					post.Comments.Set[index].Liked = true
				}
			}

			if _, okay := usersMap[comment.UserId]; okay {

				post.Comments.Set[index].User = usersMap[comment.UserId]
			}
		}

		// Sort by created at
		sort.Sort(model.ByCommentCreatedAt(post.Comments.Set))

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

	c.JSON(200, post)
}

func (di *PostAPI) PostsGetOneSlug(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	// Get the post using the slug
	slug := c.Params.ByName("slug")

	// Get the collection
	collection := database.C("posts")

	post := model.Post{}

	// Try to fetch the needed post by id
	err := collection.Find(bson.M{"slug": slug}).One(&post)

	if err != nil {

		c.JSON(404, gin.H{"error": "Couldnt found post with that slug.", "status": 203})

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

		for _, user := range users {

			if user.Id == post.UserId {

				// Set the author
				post.Author = user
			}

			usersMap[user.Id] = map[string]string{
				"id":    user.Id.Hex(),
				"name":  user.UserName,
				"email": user.Email,
			}
		}

		// Get the query parameters
		qs := c.Request.URL.Query()

		// Name of the set to get
		token := qs.Get("token")

		// Look for votes that has been already given
		var votes []model.Vote
		var likes []model.Vote

		if token != "" {

			// Get user by token
			user_token := model.UserToken{}

			// Try to fetch the user using token header though
			err = database.C("tokens").Find(bson.M{"token": token}).One(&user_token)

			if err == nil {

				err = database.C("votes").Find(bson.M{"type": "component", "related_id": post.Id, "user_id": user_token.UserId}).All(&votes)

				// Get the likes given by the current user
				_ = database.C("votes").Find(bson.M{"type": "comment", "related_id": post.Id, "user_id": user_token.UserId}).All(&likes)
			}

			if user_token.UserId != post.UserId {

				// Check if following
				following := model.UserFollowing{}

				err = database.C("followers").Find(bson.M{"follower": user_token.UserId, "following": post.UserId}).One(&following)

				// The user is following the author so tell the post struct
				if err == nil {

					post.Following = true
				}
			}

			// Increase user saw posts and its gamification in another thread
			go func(token model.UserToken, users []model.User) {

				var target model.User

				// Update the user saw posts
				_ = database.C("users").Update(bson.M{"_id": token.UserId}, bson.M{"$inc": bson.M{"stats.saw": 1}})
				player := false

				for _, user := range users {

					if user.Id == token.UserId {

						// The user is a player of the post so we dont have to get it from the database again
						player = true
						target = user
					}
				}

				if player == false {

					err = collection.Find(bson.M{"_id": token.UserId}).One(&target)

					if err != nil {
						panic(err)
					}
				}

				// Update user achievements (saw posts)
				//updateUserAchievement(target, "saw")

			}(user_token, users)
		}

		for index := range post.Comments.Set {

			comment := &post.Comments.Set[index]

			// Save the position over the comment
			post.Comments.Set[index].Position = index

			// Check if user liked that comment already
			for _, vote := range likes {

				if vote.NestedType == strconv.Itoa(index) {

					post.Comments.Set[index].Liked = true
				}
			}

			if _, okay := usersMap[comment.UserId]; okay {

				post.Comments.Set[index].User = usersMap[comment.UserId]
			}
		}

		// Sort by created at
		sort.Sort(model.ByCommentCreatedAt(post.Comments.Set))

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

	c.JSON(200, post)
}

func (di *PostAPI) PostCreate(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	// Get user by token
	user_token := model.UserToken{}
	token := c.Request.Header.Get("Auth-Token")

	if token == "" {

		c.JSON(401, gin.H{"message": "No auth credentials", "status": "error", "code": 401})
		return
	}

	// Try to fetch the user using token header though
	err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)

	if err != nil {

		c.JSON(401, gin.H{"message": "No valid auth credentials", "status": "error", "code": 401})
		return
	}

	var post model.PostForm

	// Get the form otherwise tell it has been an error
	if c.BindWith(&post, binding.JSON) == nil {

		comments := model.Comments{
			Count: 0,
			Set:   make([]model.Comment, 0),
		}

		votes := model.Votes{
			Up:     0,
			Down:   0,
			Rating: 0,
		}

		// Empty participants list - only author included
		users := []bson.ObjectId{user_token.UserId}

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
				
				post_name := "PC " + post.Name + " - presupuesto de $" + budget.(string) + " " + budget_currency.(string) 

				publish := &model.Post{
					Title:      post_name,
					Content:    post.Content,
					Type:       "recommendations",
					Slug:       sanitize.Path(post.Name),
					Comments:   comments,
					UserId:     user_token.UserId,
					Users:      users,
					Categories: []string{"recommendations"},
					Votes:      votes,
					Created:    time.Now(),
					Updated:    time.Now(),
				}
				
				publish_components := model.Components{
				    Budget: budget.(string),
				    BudgetType: budget_type.(string),
				    BudgetCurrency: budget_currency.(string),
				    BudgetFlexibility: budget_flexibility.(string),
				    Software: software.(string),
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

                        if name == component || name == component + ",omitempty" {
                            
                            c := model.Component{
                                Elections: component_elements["poll"].(bool),
                                Status:  status,
                                Votes:   votes,
                                Content: component_elements["value"].(string),
                            }
                
                            // Set the component with the component we've build above
                            bindable.Field(i).Set(reflect.ValueOf(c))
                        }
                    }
				}
                
                // Now bind the components to the post 
                publish.Components = publish_components

				err = database.C("posts").Insert(publish)

				if err != nil {
					panic(err)
				}
				
				// Add a counter for the category
				//counterAdd("recommendations")

				// Finished creating the post
				c.JSON(200, gin.H{"status": "okay", "code": 200})
				return
			}

		case "category-post":

			slug_exists, _ := database.C("posts").Find(bson.M{"slug": utils.GenerateSlug(post.Name)}).Count()
			slug := ""

			if slug_exists > 0 {

				var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

				b := make([]rune, 6)
				for i := range b {
					b[i] = letters[rand.Intn(len(letters))]
				}
				suffix := string(b)

				// Duplicated so suffix it
				slug = sanitize.Path(post.Name) + "-" + suffix

			} else {

				// No duplicates
				slug = utils.GenerateSlug(post.Name)
			}

			publish := &model.Post{
				Title:      post.Name,
				Content:    post.Content,
				Type:       "category-post",
				Slug:       slug,
				Comments:   comments,
				UserId:     user_token.UserId,
				Users:      users,
				Categories: []string{post.Tag},
				Votes:      votes,
				Created:    time.Now(),
				Updated:    time.Now(),
			}

			err = database.C("posts").Insert(publish)

			if err != nil {
				panic(err)
			}
			
			// Add a counter for the category
			//counterAdd(post.Tag)

			// Finished creating the post
			c.JSON(200, gin.H{"status": "okay", "code": 200})
			return
		}
	}

	c.JSON(400, gin.H{"status": "error", "message": "Couldnt create post, missing information...", "code": 205})
}

func (di *PostAPI) resetUserCategoryCounter(category string, user_id bson.ObjectId) {
    
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