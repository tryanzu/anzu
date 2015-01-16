package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/mrvdot/golang-utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

/**
 * Avaliable components
 *
 * List of valid components along a recommendation post
 */
var avaliable_components = []string{
	"cpu", "motherboard", "ram", "cabinet", "screen", "storage", "cooler", "power", "videocard",
}

type Votes struct {
	Up     int `bson:"up" json:"up"`
	Down   int `bson:"down" json:"down"`
	Rating int `bson:"rating,omitempty" json:"rating,omitempty"`
}

type Author struct {
	Id      bson.ObjectId `bson:"id,omitempty" json:"id,omitempty"`
	Name    string        `bson:"name" json:"name"`
	Email   string        `bson:"email" json:"email"`
	Avatar  string        `bson:"avatar" json:"avatar"`
	Profile interface{}   `bson:"profile,omitempty" json:"profile,omitempty"`
}

type Comments struct {
	Count int       `bson:"count" json:"count"`
	Set   []Comment `bson:"set" json:"set"`
}

type Comment struct {
	UserId   bson.ObjectId `bson:"user_id" json:"user_id"`
	Votes    Votes         `bson:"votes" json:"votes"`
	User     interface{}   `bson:"author,omitempty" json:"author,omitempty"`
	Position int           `bson:"position,omitempty" json:"position"`
	Liked    bool          `bson:"liked,omitempty" json:"liked,omitempty"`
	Content  string        `bson:"content" json:"content"`
	Created  time.Time     `bson:"created_at" json:"created_at"`
}

type Components struct {
	Cpu               Component `bson:"cpu,omitempty" json:"cpu,omitempty"`
	Motherboard       Component `bson:"motherboard,omitempty" json:"motherboard,omitempty"`
	Ram               Component `bson:"ram,omitempty" json:"ram,omitempty"`
	Storage           Component `bson:"storage,omitempty" json:"storage,omitempty"`
	Cooler            Component `bson:"cooler,omitempty" json:"cooler,omitempty"`
	Power             Component `bson:"power,omitempty" json:"power,omitempty"`
	Cabinet           Component `bson:"cabinet,omitempty" json:"cabinet,omitempty"`
	Screen            Component `bson:"screen,omitempty" json:"screen,omitempty"`
	Videocard         Component `bson:"videocard,omitempty" json:"videocard,omitempty"`
	Software          string    `bson:"software,omitempty" json:"software,omitempty"`
	Budget            string    `bson:"budget,omitempty" json:"budget,omitempty"`
	BudgetCurrency    string    `bson:"budget_currency,omitempty" json:"budget_currency,omitempty"`
	BudgetType        string    `bson:"budget_type,omitempty" json:"budget_type,omitempty"`
	BudgetFlexibility string    `bson:"budget_flexibility,omitempty" json:"budget_flexibility,omitempty"`
}

type Component struct {
	Content   string           `bson:"content" json:"content"`
	Elections bool             `bson:"elections" json:"elections"`
	Options   []ElectionOption `bson:"options,omitempty" json:"options"`
	Votes     Votes            `bson:"votes" json:"votes"`
	Status    string           `bson:"status" json:"status"`
	Voted     string           `bson:"voted,omitempty" json:"voted,omitempty"`
}

type Post struct {
	Id         bson.ObjectId   `bson:"_id,omitempty" json:"id,omitempty"`
	Title      string          `bson:"title" json:"title"`
	Slug       string          `bson:"slug" json:"slug"`
	Type       string          `bson:"type" json:"type"`
	Content    string          `bson:"content" json:"content"`
	Categories []string        `bson:"categories" json:"categories"`
	Comments   Comments        `bson:"comments" json:"comments"`
	Author     User            `bson:"author,omitempty" json:"author,omitempty"`
	UserId     bson.ObjectId   `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Users      []bson.ObjectId `bson:"users,omitempty" json:"users,omitempty"`
	Votes      Votes           `bson:"votes" json:"votes"`
	Components Components      `bson:"components,omitempty" json:"components,omitempty"`
	Following  bool            `bson:"following,omitempty" json:"following,omitempty"`
	Created    time.Time       `bson:"created_at" json:"created_at"`
	Updated    time.Time       `bson:"updated_at" json:"updated_at"`
}

type PostForm struct {
	Kind       string                 `json:"kind" binding:"required"`
	Name       string                 `json:"name" binding:"required"`
	Content    string                 `json:"content" binding:"required"`
	Budget     string                 `json:"budget"`
	Currency   string                 `json:"currency"`
	Moves      string                 `json:"moves"`
	Software   string                 `json:"software"`
	Tag        string                 `json:"tag"`
	Components map[string]interface{} `json:"components"`
}

// ByCommentCreatedAt implements sort.Interface for []ElectionOption based on Created field
type ByCommentCreatedAt []Comment

func (a ByCommentCreatedAt) Len() int           { return len(a) }
func (a ByCommentCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCommentCreatedAt) Less(i, j int) bool { return !a[i].Created.Before(a[j].Created) }

func FeedGet(c *gin.Context) {

	var feed []Post
	offset := 0
	limit := 10

	qs := c.Request.URL.Query()

	o := qs.Get("offset")
	l := qs.Get("limit")

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

	// Prepare the database to fetch the feed
	posts_collection := database.C("posts")
	get_feed := posts_collection.Find(bson.M{}).Sort("-created_at").Limit(limit).Skip(offset)

	// Get the results from the feed algo
	err := get_feed.All(&feed)

	if err != nil {
		panic(err)
	}

	var authors []bson.ObjectId
	var users []User

	for _, post := range feed {

		authors = append(authors, post.UserId)
	}

	// Get the users needed by the feed
	err = database.C("users").Find(bson.M{"_id": bson.M{"$in": authors}}).All(&users)

	if err != nil {
		panic(err)
	}

	if len(feed) > 0 {

		usersMap := make(map[bson.ObjectId]User)

		for _, user := range users {

			usersMap[user.Id] = user
		}

		for index := range feed {

			post := &feed[index]

			if _, okay := usersMap[post.UserId]; okay {

				postUser := usersMap[post.UserId]

				post.Author = User{
					Id:        postUser.Id,
					UserName:  postUser.UserName,
					FirstName: postUser.FirstName,
					LastName:  postUser.LastName,
					Email:     postUser.Email,
				}
			}
		}

		c.JSON(200, gin.H{"feed": feed, "offset": offset, "limit": limit})
	} else {

		c.JSON(200, gin.H{"feed": []string{}, "offset": offset, "limit": limit})
	}
}

func PostsGet(c *gin.Context) {

	var recommendations []Post
	var published []Post

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

	var users []User

	// Get the users
	collection := database.C("users")

	err = collection.Find(bson.M{"_id": bson.M{"$in": authors}}).All(&users)

	if err != nil {
		panic(err)
	}

	usersMap := make(map[bson.ObjectId]User)

	for _, user := range users {

		usersMap[user.Id] = user
	}

	for index := range recommendations {

		post := &recommendations[index]

		if _, okay := usersMap[post.UserId]; okay {

			postUser := usersMap[post.UserId]

			post.Author = User{
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

			post.Author = User{
				Id:        postUser.Id,
				UserName:  postUser.UserName,
				FirstName: postUser.FirstName,
				LastName:  postUser.LastName,
			}
		}
	}

	c.JSON(200, gin.H{"recommendations": recommendations, "last": published})
}

func PostsGetOne(r render.Render, database *mgo.Database, req *http.Request, params martini.Params) {

	if bson.IsObjectIdHex(params["id"]) == false {

		response := map[string]string{
			"error":  "Invalid params to get a post.",
			"status": "202",
		}

		r.JSON(400, response)

		return
	}

	// Get the id of the needed post
	id := bson.ObjectIdHex(params["id"])

	// Get the collection
	collection := database.C("posts")

	post := Post{}

	// Try to fetch the needed post by id
	err := collection.FindId(id).One(&post)

	if err != nil {

		response := map[string]string{
			"error":  "Couldnt found post with that id.",
			"status": "201",
		}

		r.JSON(404, response)

		return
	}

	r.JSON(200, post)
}

func PostsGetOneSlug(c *gin.Context) {

	// Get the post using the slug
	slug := c.Params.ByName("slug")

	// Get the collection
	collection := database.C("posts")

	post := Post{}

	// Try to fetch the needed post by id
	err := collection.Find(bson.M{"slug": slug}).One(&post)

	if err != nil {

		c.JSON(404, gin.H{"error": "Couldnt found post with that slug.", "status": 203})

		return
	}

	// Get the users and stuff
	if post.Users != nil && len(post.Users) > 0 {

		var users []User

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
		var votes []Vote
		var likes []Vote

		if token != "" {

			// Get user by token
			user_token := UserToken{}

			// Try to fetch the user using token header though
			err = database.C("tokens").Find(bson.M{"token": token}).One(&user_token)

			if err == nil {

				err = database.C("votes").Find(bson.M{"type": "component", "related_id": post.Id, "user_id": user_token.UserId}).All(&votes)

				// Get the likes given by the current user
				_ = database.C("votes").Find(bson.M{"type": "comment", "related_id": post.Id, "user_id": user_token.UserId}).All(&likes)
			}

			if user_token.UserId != post.UserId {

				// Check if following
				following := UserFollowing{}

				err = database.C("followers").Find(bson.M{"follower": user_token.UserId, "following": post.UserId}).One(&following)

				// The user is following the author so tell the post struct
				if err == nil {

					post.Following = true
				}
			}

			// Increase user saw posts and its gamification in another thread
			go func(token UserToken, users []User) {

				var target User

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
				updateUserAchievement(target, "saw")

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
		sort.Sort(ByCommentCreatedAt(post.Comments.Set))

		components := reflect.ValueOf(&post.Components).Elem()
		components_type := reflect.TypeOf(&post.Components).Elem()

		for i := 0; i < components.NumField(); i++ {

			f := components.Field(i)
			t := components_type.Field(i)

			if f.Type().String() == "main.Component" {

				component := f.Interface().(Component)

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
					sort.Sort(ByElectionsCreatedAt(component.Options))
				}

				f.Set(reflect.ValueOf(component))
			}
		}
	}

	c.JSON(200, post)
}

func PostCreate(c *gin.Context) {

	// Get user by token
	user_token := UserToken{}
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

	var post PostForm

	// Get the form otherwise tell it has been an error
	if c.BindWith(&post, binding.JSON) {

		comments := Comments{
			Count: 0,
			Set:   make([]Comment, 0),
		}

		votes := Votes{
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

				publish := &Post{
					Title:      post.Name,
					Content:    post.Content,
					Type:       "category-post",
					Slug:       utils.GenerateSlug(post.Name),
					Comments:   comments,
					UserId:     user_token.UserId,
					Users:      users,
					Categories: []string{"el-bar"},
					Votes:      votes,
					Created:    time.Now(),
					Updated:    time.Now(),
				}
				
				publish_components := Components{
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
                            
                            c := Component{
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
				slug = utils.GenerateSlug(post.Name) + "-" + suffix

			} else {

				// No duplicates
				slug = utils.GenerateSlug(post.Name)
			}

			publish := &Post{
				Title:      post.Name,
				Content:    post.Content,
				Type:       "category-post",
				Slug:       slug,
				Comments:   comments,
				UserId:     user_token.UserId,
				Users:      users,
				Categories: []string{"el-bar"},
				Votes:      votes,
				Created:    time.Now(),
				Updated:    time.Now(),
			}

			err = database.C("posts").Insert(publish)

			if err != nil {
				panic(err)
			}

			// Finished creating the post
			c.JSON(200, gin.H{"status": "okay", "code": 200})
			return
		}
	}

	c.JSON(400, gin.H{"status": "error", "message": "Couldnt create post, missing information...", "code": 205})
}
