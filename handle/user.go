package handle

import (
	"code.google.com/p/go-uuid/uuid"
	"crypto/sha256"
	"encoding/hex"
	"github.com/CloudCom/fireauth"
	"github.com/dgrijalva/jwt-go"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/kennygrant/sanitize"
	"github.com/mitchellh/goamz/s3"
	"github.com/mrvdot/golang-utils"
	"github.com/olebedev/config"
	"github.com/xuyu/goredis"
	"gopkg.in/h2non/bimg.v0"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

type UserAPI struct {
	DataService   *mongo.Service `inject:""`
	CacheService  *goredis.Redis `inject:""`
	ConfigService *config.Config `inject:""`
	S3Bucket      *s3.Bucket     `inject:""`
	Collector     CollectorAPI   `inject:"inline"`
}

func (di *UserAPI) UserSubscribe(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	var register model.UserSubscribeForm

	if c.BindWith(&register, binding.JSON) == nil {

		subscribe := &model.UserSubscribe{
			Category: register.Category,
			Email:    register.Email,
		}

		err := database.C("subscribes").Insert(subscribe)

		if err != nil {
			panic(err)
		}
		c.JSON(200, gin.H{"status": "okay"})
	}
}


func (di *UserAPI) UserCategorySubscribe(c *gin.Context) {

	var user model.User

	// Get the database interface from the DI
	database 	 := di.DataService.Database
	redis := di.CacheService
	user_id  	 := c.MustGet("user_id")
	category_id  := c.Param("id")
	user_bson_id := bson.ObjectIdHex(user_id.(string))

	if bson.IsObjectIdHex(category_id) == false {

		c.JSON(400, gin.H{"status": "error", "message": "Invalid category id."})
		return
	}

	_, err := database.C("categories").Find(bson.M{"_id": bson.ObjectIdHex(category_id), "parent": bson.M{"$exists": true}}).Count()

	if err != nil {

		c.JSON(400, gin.H{"status": "error", "message": "No such category."})
		return
	}

	// Get the user using the session
	err = database.C("users").Find(bson.M{"_id": user_bson_id}).One(&user)

	if err != nil {
		panic(err)
	}
	
	for _, user_category_id := range user.Categories {

		if user_category_id.Hex() == category_id {

			c.JSON(200, gin.H{"status": "okay", "message": "already-following"})
			return
		}
	}

	err = database.C("users").Update(bson.M{"_id": user.Id}, bson.M{"$push": bson.M{"categories": bson.ObjectIdHex(category_id)}})

	if err != nil {
		panic(err)
	}

	// Create the set inside redis and move on
	redis.SAdd("user:categories:" + user_id.(string), category_id) 

	c.JSON(200, gin.H{"status": "okay"})
}

func (di *UserAPI) UserCategoryUnsubscribe(c *gin.Context) {

	// Get the database interface from the DI
	database 	 := di.DataService.Database
	redis := di.CacheService
	user_id  	 := c.MustGet("user_id")
	category_id  := c.Param("id")
	user_bson_id := bson.ObjectIdHex(user_id.(string))

	if bson.IsObjectIdHex(category_id) == false {

		c.JSON(400, gin.H{"status": "error", "message": "Invalid category id."})
		return
	}

	_, err := database.C("categories").Find(bson.M{"_id": bson.ObjectIdHex(category_id), "parent": bson.M{"$exists": true}}).Count()

	if err != nil {

		c.JSON(400, gin.H{"status": "error", "message": "No such category."})
		return
	}

	remove := []bson.ObjectId{bson.ObjectIdHex(category_id)}
	err = database.C("users").Update(bson.M{"_id": user_bson_id}, bson.M{"$pullAll": bson.M{"categories": remove}})

	if err != nil {
		panic(err)
	}

	// Create the set inside redis and move on
	redis.SRem("user:categories:" + user_id.(string), category_id) 

	c.JSON(200, gin.H{"status": "okay"})
}

func (di *UserAPI) UserGetOne(c *gin.Context) {

	database := di.DataService.Database
	user_id := c.Param("id")

	if bson.IsObjectIdHex(user_id) == false {

		c.JSON(400, gin.H{"status": "error", "message": "Invalid user id."})
		return
	}

	user_bson_id := bson.ObjectIdHex(user_id)

	// Get the user using the specified id
	user := model.User{}
	err := database.C("users").Find(bson.M{"_id": user_bson_id}).One(&user)

	if err != nil {
		panic(err)
	}

	// Save the activity
	user_logged_id, signed_in := c.Get("user_id")

	if signed_in {

		// Save the activity in other routine
		go di.Collector.Activity(model.Activity{UserId: bson.ObjectIdHex(user_logged_id.(string)), Event: "user", RelatedId: user.Id})
	}

	c.JSON(200, user)
}

func (di *UserAPI) UserGetByToken(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database
	user_id := c.MustGet("user_id")
	user_bson_id := bson.ObjectIdHex(user_id.(string))

	// Get the user using the session
	user := model.User{}
	err := database.C("users").Find(bson.M{"_id": user_bson_id}).One(&user)

	if err != nil {
		panic(err)
	}

	// Get the user notifications
	notifications, err := database.C("notifications").Find(bson.M{"user_id": user.Id, "seen": false}).Count()

	if err != nil {
		panic(err)
	}

	user.Notifications = notifications

	// Alright, go back and send the user info
	c.JSON(200, user)
}

func (di *UserAPI) UserGetToken(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	// Get the query parameters
	qs := c.Request.URL.Query()

	// Get the email or the username or the id and its password
	email, password := qs.Get("email"), qs.Get("password")

	collection := database.C("users")

	user := model.User{}

	// Try to fetch the user using email param though
	err := collection.Find(bson.M{"email": email}).One(&user)

	if err != nil {

		c.JSON(400, gin.H{"status": "error", "message": "Couldnt found user with that email", "code": 400})
		return
	}

	// Incorrect password
	password_encrypted := []byte(password)
	sha256 := sha256.New()
	sha256.Write(password_encrypted)
	md := sha256.Sum(nil)
	hash := hex.EncodeToString(md)

	if user.Password != hash {

		c.JSON(400, gin.H{"status": "error", "message": "Credentials are not correct", "code": 400})
		return
	}

	// Generate user token
	uuid := uuid.New()
	token := &model.UserToken{
		UserId:  user.Id,
		Token:   uuid,
		Closed:  false,
		Created: time.Now(),
		Updated: time.Now(),
	}

	err = database.C("tokens").Insert(token)

	c.JSON(200, token)
}

func (di *UserAPI) UserGetJwtToken(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	// Get the query parameters
	qs := c.Request.URL.Query()

	// Get the email or the username or the id and its password
	email, password := qs.Get("email"), qs.Get("password")
	collection := database.C("users")
	user := model.User{}

	// Try to fetch the user using email param though
	err := collection.Find(bson.M{"email": email}).One(&user)

	if err != nil {

		c.JSON(400, gin.H{"status": "error", "message": "Couldnt found user with that email", "code": 400})
		return
	}

	// Check whether the password match the user password or not
	password_encrypted := []byte(password)
	sha256 := sha256.New()
	sha256.Write(password_encrypted)
	md := sha256.Sum(nil)
	hash := hex.EncodeToString(md)

	if user.Password != hash {
		c.JSON(400, gin.H{"status": "error", "message": "Credentials are not correct", "code": 400})
		return
	}

	// Generate JWT with the information about the user
	token, firebase := di.generateUserToken(user.Id)

	// Save the activity
	user_id, signed_in := c.Get("user_id")

	if signed_in {

		// Save the activity in other routine
		go di.Collector.Activity(model.Activity{UserId: bson.ObjectIdHex(user_id.(string)), Event: "user-view", RelatedId: user.Id})
	}

	c.JSON(200, gin.H{"status": "okay", "token": token, "firebase": firebase})
}

func (di *UserAPI) UserGetTokenFacebook(c *gin.Context) {

	var facebook map[string]interface{}
	var id bson.ObjectId

	// Get the database interface from the DI
	database := di.DataService.Database

	// Bind to strings map
	c.BindWith(&facebook, binding.JSON)

	var facebook_id interface{}

	if _, okay := facebook["id"]; okay == false {

		c.JSON(401, gin.H{"error": "Invalid oAuth facebook token...", "status": 105})
		return
	} else {

		facebook_id = facebook["id"]
	}

	collection := database.C("users")
	user := model.User{}

	// Try to fetch the user using the facebook id param though
	err := collection.Find(bson.M{"facebook.id": facebook_id}).One(&user)

	// Create a new user
	if err != nil {

		var facebook_first_name, facebook_last_name, facebook_email string

		username := facebook["first_name"].(string) + " " + facebook["last_name"].(string)
		id = bson.NewObjectId()

		// Ensure the facebook data is alright
		if _, ok := facebook["first_name"]; ok {

			facebook_first_name = facebook["first_name"].(string)
		} else {

			facebook_first_name = ""
		}

		if _, ok := facebook["last_name"]; ok {

			facebook_last_name = facebook["last_name"].(string)
		} else {

			facebook_last_name = ""
		}

		if _, ok := facebook["email"]; ok {

			facebook_email = facebook["email"].(string)
		} else {

			facebook_email = ""
		}

		user := &model.User{
			Id:          id,
			FirstName:   facebook_first_name,
			LastName:    facebook_last_name,
			UserName:    utils.GenerateSlug(username),
			Password:    "",
			Email:       facebook_email,
			Roles:       make([]string, 0),
			Permissions: make([]string, 0),
			NameChanges: 0,
			Description: "",
			Facebook:    facebook,
			Created:     time.Now(),
			Updated:     time.Now(),
		}

		err = database.C("users").Insert(user)

		if err != nil {

			c.JSON(500, gin.H{"error": "Could create user using facebook oAuth...", "status": 106})
			return
		}

		err = database.C("counters").Insert(model.Counter{
			UserId: id,
			Counters: map[string]model.PostCounter{
				"news": model.PostCounter{
					Counter: 0,
					Updated: time.Now(),
				},
			},
		})

	} else {

		// The id for the token would be the same as the facebook user
		id = user.Id
	}

	// Generate JWT with the information about the user
	token, firebase := di.generateUserToken(id)

	c.JSON(200, gin.H{"status": "okay", "token": token, "firebase": firebase})
}

func (di *UserAPI) UserUpdateProfileAvatar(c *gin.Context) {

	// Check for user token
	user_id := c.MustGet("user_id")
	user_bson_id := bson.ObjectIdHex(user_id.(string))

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

		path := "users/" + name + extension
		err = di.S3Bucket.Put(path, data, dataType, s3.ACL("public-read"))

		if err != nil {
			panic(err)
		}

		options := bimg.Options{
			Width:  120,
			Height: 120,
			Embed:  true,
			Crop:   true,
			Quality: 100,
		}

		thumbnail, err := bimg.NewImage(data).Process(options)

		if err != nil {
			panic(err)
		}

		path = "users/" + name + "-120x120" + extension
		err = di.S3Bucket.Put(path, thumbnail, dataType, s3.ACL("public-read"))

		if err != nil {
			panic(err)
		}

		s3_url := "http://s3-us-west-1.amazonaws.com/spartan-board/" + path

		// Update the user image as well
		di.DataService.Database.C("users").Update(bson.M{"_id": user_bson_id}, bson.M{"$set": bson.M{"image": s3_url}})

		// Done
		c.JSON(200, gin.H{"status": "okay", "url": s3_url})

		return
	}

	c.JSON(400, gin.H{"status": "error", "message": "Could not detect an image file..."})
}

func (di *UserAPI) UserUpdateProfile(c *gin.Context) {

	var user model.User

	// Get the database interface from the DI
	database := di.DataService.Database

	// Users collection
	users_collection := database.C("users")

	// Check for user token
	user_id := c.MustGet("user_id")
	user_bson_id := bson.ObjectIdHex(user_id.(string))

	err := users_collection.Find(bson.M{"_id": user_bson_id}).One(&user)

	if err == nil {

		var profileUpdate model.UserProfileForm

		if c.BindWith(&profileUpdate, binding.JSON) == nil {

			set := bson.M{}

			if profileUpdate.UserName != "" && user.NameChanges < 1 {

				valid_username, _ := regexp.Compile(`^[0-9a-zA-Z\-]{0,32}$`)

				if valid_username.MatchString(profileUpdate.UserName) {

					// Generate a slug for the username
					username_slug := sanitize.Path(sanitize.Accents(profileUpdate.UserName))

					// Check whether user exists
					count, _ := database.C("users").Find(bson.M{"username_slug": username_slug}).Count()

					if count == 0 {

						set["username"] = profileUpdate.UserName
						set["username_slug"] = username_slug
						set["name_changes"] = user.NameChanges + 1
					}
				}
			}

			set["updated_at"] = time.Now()

			// Update the user profile with some godness
			users_collection.Update(bson.M{"_id": user.Id}, bson.M{"$set": set})

			c.JSON(200, gin.H{"message": "okay", "status": "okay", "code": 200})
			return
		}
	}

	c.JSON(400, gin.H{"status": "error", "message": "Invalid auth request."})
}

func (di *UserAPI) UserRegisterAction(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	var registerAction model.UserRegisterForm

	if c.BindWith(&registerAction, binding.JSON) == nil {

		// Check if already registered
		email_exists, _ := database.C("users").Find(bson.M{"email": registerAction.Email}).Count()

		if email_exists > 0 {

			// Only one account per email
			c.JSON(400, gin.H{"status": "error", "message": "User already registered", "code": 470})
			return
		}

		valid_username, _ := regexp.Compile(`^[0-9a-zA-Z\-]{0,32}$`)

		if valid_username.MatchString(registerAction.UserName) == false {

			// Only some characters in the username
			c.JSON(400, gin.H{"status": "error", "message": "Username not valid"})
			return
		}

		username_slug := sanitize.Path(sanitize.Accents(registerAction.UserName))
		user_exists, _ := database.C("users").Find(bson.M{"username_slug": username_slug}).Count()

		if user_exists > 0 {

			// Username busy
			c.JSON(400, gin.H{"status": "error", "message": "User already registered", "code": 471})
			return
		}

		// Encode password using sha
		password_encrypted := []byte(registerAction.Password)
		sha256 := sha256.New()
		sha256.Write(password_encrypted)
		md := sha256.Sum(nil)
		hash := hex.EncodeToString(md)

		// Profile default data
		profile := gin.H{
			"step":           0,
			"ranking":        0,
			"country":        "México",
			"posts":          0,
			"followers":      0,
			"show_email":     false,
			"favourite_game": "-",
			"microsoft":      "-",
			"bio":            "Just another spartan geek",
		}

		id := bson.NewObjectId()

		user := &model.User{
			Id:           id,
			FirstName:    "",
			LastName:     "",
			UserName:     registerAction.UserName,
			UserNameSlug: username_slug,
			NameChanges:  1,
			Password:     hash,
			Email:        registerAction.Email,
			Roles:        []string{"registered"},
			Permissions:  make([]string, 0),
			Description:  "",
			Profile:      profile,
			Stats: model.UserStats{
				Saw: 0,
			},
			Created: time.Now(),
			Updated: time.Now(),
		}

		err := database.C("users").Insert(user)

		if err != nil {
			panic(err)
		}

		err = database.C("counters").Insert(model.Counter{
			UserId: id,
			Counters: map[string]model.PostCounter{
				"news": model.PostCounter{
					Counter: 0,
					Updated: time.Now(),
				},
			},
		})

		// Generate token if auth is going to perform
		token, firebase := di.generateUserToken(user.Id)

		// Finished creating the post
		c.JSON(200, gin.H{"status": "okay", "code": 200, "token": token, "firebase": firebase})
		return
	}

	// Couldn't process the request though
	c.JSON(400, gin.H{"status": "error", "message": "Missing information to process the request", "code": 400})
}

func (di *UserAPI) UserInvolvedFeedGet(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	var user_posts []model.Post
	var commented_posts []model.Post
	var activity = make([]model.UserActivity, 0)

	// Check whether auth or not
	user_token := model.UserToken{}
	token := c.Request.Header.Get("Auth-Token")

	if token != "" {

		// Try to fetch the user using token header though
		err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)

		if err == nil {

			var user model.User

			// Get the current user
			err := database.C("users").Find(bson.M{"_id": user_token.UserId}).One(&user)

			if err != nil {
				panic(err)
			}

			// Get the user owned posts
			err = database.C("posts").Find(bson.M{"user_id": user_token.UserId}).All(&user_posts)

			if err != nil {
				panic(err)
			}

			// Get the posts where the user commented
			err = database.C("posts").Find(bson.M{"users": user_token.UserId, "user_id": bson.M{"$ne": user_token.UserId}}).All(&commented_posts)

			if err != nil {
				panic(err)
			}

			for _, post := range user_posts {

				activity = append(activity, model.UserActivity{
					Title:     post.Title,
					Content:   post.Content,
					Created:   post.Created,
					Directive: "owner",
					Author: map[string]string{
						"id":    user.Id.Hex(),
						"name":  user.UserName,
						"email": user.Email,
					},
				})
			}

			for _, post := range commented_posts {

				for _, comment := range post.Comments.Set {

					if comment.UserId == user.Id {

						activity = append(activity, model.UserActivity{
							Title:     post.Title,
							Content:   comment.Content,
							Created:   comment.Created,
							Directive: "commented",
							Author: map[string]string{
								"id":    user.Id.Hex(),
								"name":  user.UserName,
								"email": user.Email,
							},
						})
					}
				}
			}

			// Sort the full set of posts by the time they happened
			sort.Sort(model.ByCreatedAt(activity))

			c.JSON(200, gin.H{"activity": activity})
		}
	}
}

func (di *UserAPI) UserAutocompleteGet(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	var users []gin.H

	qs := c.Request.URL.Query()
	name := qs.Get("search")

	if name != "" {

		err := database.C("users").Find(bson.M{"username": bson.RegEx{"^" + name, "i"}}).Select(bson.M{"_id": 1, "username": 1, "email": 1}).All(&users)

		if err != nil {
			panic(err)
		}

		c.JSON(200, gin.H{"users": users})
	}
}

func (di *UserAPI) generateUserToken(id bson.ObjectId) (string, string) {

	// Generate JWT with the information about the user
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims["user_id"] = id.Hex()
	token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Use the secret inside the configuration to encrypt it
	secret, err := di.ConfigService.String("application.secret")
	if err != nil {
		panic(err)
	}

	firebase_secret, err := di.ConfigService.String("firebase.secret")
	if err != nil {
		panic(err)
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}

	// Generate firebase auth token for further usage
	firebase_auth := fireauth.New(firebase_secret)
	firebase_data := fireauth.Data{"uid": id.Hex()}

	firebase_token, err := firebase_auth.CreateToken(firebase_data, nil)
	if err != nil {
		panic(err)
	}

	return tokenString, firebase_token
}
