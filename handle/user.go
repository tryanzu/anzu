package handle

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/CloudCom/fireauth"
	"github.com/dgrijalva/jwt-go"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/modules/security"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/fernandez14/go-siftscience"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/kennygrant/sanitize"
	"github.com/mitchellh/goamz/s3"
	"github.com/olebedev/config"
	"github.com/xuyu/goredis"
	"gopkg.in/h2non/bimg.v0"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"
)

type UserAPI struct {
	Errors        *exceptions.ExceptionsModule `inject:""`
	DataService   *mongo.Service   `inject:""`
	CacheService  *goredis.Redis   `inject:""`
	ConfigService *config.Config   `inject:""`
	S3Bucket      *s3.Bucket       `inject:""`
	User          *user.Module     `inject:""`
	Gaming        *gaming.Module   `inject:""`
	Security      *security.Module `inject:""`
	Collector     CollectorAPI     `inject:"inline"`
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
	database := di.DataService.Database
	redis := di.CacheService
	user_id := c.MustGet("user_id")
	category_id := c.Param("id")
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
	redis.SAdd("user:categories:"+user_id.(string), category_id)

	c.JSON(200, gin.H{"status": "okay"})
}

func (di *UserAPI) UserCategoryUnsubscribe(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database
	redis := di.CacheService
	user_id := c.MustGet("user_id")
	category_id := c.Param("id")
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
	redis.SRem("user:categories:"+user_id.(string), category_id)

	c.JSON(200, gin.H{"status": "okay"})
}

func (di *UserAPI) UserGetOne(c *gin.Context) {

	user_id := c.Param("id")

	if bson.IsObjectIdHex(user_id) == false {

		c.JSON(400, gin.H{"status": "error", "message": "Invalid user id."})
		return
	}

	user_bson_id := bson.ObjectIdHex(user_id)

	// Get the user using its id
	usr, err := di.User.Get(user_bson_id)

	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	usr.Load("components")

	// Save the activity
	user_logged_id, signed_in := c.Get("user_id")

	if signed_in {

		// Save the activity in other routine
		go di.Collector.Activity(model.Activity{UserId: bson.ObjectIdHex(user_logged_id.(string)), Event: "user", RelatedId: usr.Data().Id})
	}

	c.JSON(200, usr.Load("referrals").Data().User)
}

func (di *UserAPI) UserGetByToken(c *gin.Context) {

	id := c.MustGet("user_id")

	if bson.IsObjectIdHex(id.(string)) == false {

		// Dont allow the request
		c.JSON(400, gin.H{"status": "error", "message": "Invalid request, need valid token."})
		return
	}

	user_id := bson.ObjectIdHex(id.(string))

	// Get the user using its id
	usr, err := di.User.Get(user_id)

	if err != nil {

		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	trusted := di.Security.TrustUserIP(c.ClientIP(), usr)

	if !trusted {
		c.JSON(403, gin.H{"status": "error", "message": "Not trusted."})
		return
	}

	go func(usr *user.One) {

		// Track user sign in
		usr.TrackUserSignin(c.ClientIP())

		// Does daily login calculations
		di.Gaming.Get(usr).DailyLogin()

	}(usr)

	session_id := c.MustGet("session_id").(string)

	data := usr.Data()
	data.SessionId = session_id

	// Alright, go back and send the user info
	c.JSON(200, data)
}

func (di UserAPI) UserGetJwtToken(c *gin.Context) {

	trusted := di.Security.TrustIP(c.ClientIP())

	if !trusted {

		c.JSON(403, gin.H{"status": "error", "message": "Not trusted."})
		return
	}

	qs := c.Request.URL.Query()

	// Get the email or the username or the id and its password
	email, password := qs.Get("email"), qs.Get("password")
	usr, err := di.User.Get(bson.M{"email": email})

	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Couldnt get user."})
		return
	}

	// Check whether the password match the user password or not
	password_encrypted := []byte(password)
	sha256 := sha256.New()
	sha256.Write(password_encrypted)
	md := sha256.Sum(nil)
	hash := hex.EncodeToString(md)

	if usr.Data().Password != hash {
		c.JSON(400, gin.H{"status": "error", "message": "Credentials are not correct", "code": 400})
		return
	}

	trusted_user := di.Security.TrustUserIP(c.ClientIP(), usr)

	if !trusted_user {

		c.JSON(403, gin.H{"status": "error", "message": "Not trusted."})
		return
	}

	session_id := c.MustGet("session_id").(string)

	go di.trackSiftScienceLogin(usr.Data().Id.Hex(), session_id, true)

	// Generate JWT with the information about the user
	token, firebase := di.generateUserToken(usr.Data().Id)

	// Save the activity
	user_id, signed_in := c.Get("user_id")

	if signed_in {

		// Save the activity in other routine
		go di.Collector.Activity(model.Activity{UserId: bson.ObjectIdHex(user_id.(string)), Event: "user-view", RelatedId: usr.Data().Id})
	}

	c.JSON(200, gin.H{"status": "okay", "token": token, "session_id": session_id, "firebase": firebase})
}

func (di UserAPI) UserGetTokenFacebook(c *gin.Context) {

	var facebook map[string]interface{}
	var id bson.ObjectId

	// Bind to strings map
	c.BindWith(&facebook, binding.JSON)

	var facebook_id interface{}

	if _, okay := facebook["id"]; okay == false {

		c.JSON(401, gin.H{"error": "Invalid oAuth facebook token...", "status": 105})
		return
	} else {

		facebook_id = facebook["id"]
	}

	session_id := c.MustGet("session_id").(string)
	usr, err := di.User.Get(bson.M{"facebook.id": facebook_id})

	// Create a new user
	if err != nil {

		trusted := di.Security.TrustIP(c.ClientIP())

		if !trusted {

			c.JSON(403, gin.H{"status": "error", "message": "Not trusted."})
			return
		}

		_, err := di.User.SignUpFacebook(facebook)

		if err != nil {

			c.JSON(400, gin.H{"status": "error", "message": err.Error()})
			return
		}

	} else {

		// The id for the token would be the same as the facebook user
		id = usr.Data().Id
		trusted_user := di.Security.TrustUserIP(c.ClientIP(), usr)

		if !trusted_user {
			c.JSON(403, gin.H{"status": "error", "message": "Not trusted."})
			return
		}

		if email, exists := facebook["email"]; exists {
			_ = usr.Update(map[string]interface{}{"facebook": facebook, "email": email.(string)})
		}

		go di.trackSiftScienceLogin(id.Hex(), session_id, true)
	}

	// Generate JWT with the information about the user
	token, firebase := di.generateUserToken(id)

	c.JSON(200, gin.H{"status": "okay", "token": token, "firebase": firebase, "session_id": session_id})
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
			Width:   120,
			Height:  120,
			Embed:   true,
			Crop:    true,
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

		s3_url := "https://s3-us-west-1.amazonaws.com/spartan-board/" + path

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

			if profileUpdate.Description != "" {

				description := profileUpdate.Description

				if len([]rune(description)) > 60 {
					description = helpers.Truncate(description, 57) + "..."
				}

				set["description"] = description
			}

			if profileUpdate.Password != "" {

				if len([]rune(profileUpdate.Password)) < 4 {
					c.JSON(400, gin.H{"status": "error", "message": "Can't allow password update, too short."})
					return
				}

				password := helpers.Sha256(profileUpdate.Password)

				set["password"] = password
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

func (di *UserAPI) UserValidateEmail(c *gin.Context) {

	code := c.Param("code")

	// Attempt to get the user by the confirmation code
	usr, err := di.User.Get(bson.M{"ver_code": code})

	if err != nil {

		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	usr.MarkAsValidated()

	c.JSON(200, gin.H{"status": "okay"})
}

func (di *UserAPI) UserRegisterAction(c *gin.Context) {

	var form model.UserRegisterForm

	if c.BindWith(&form, binding.JSON) == nil {


		// Get the user using its id
		usr, err := di.User.SignUp(form.Email, form.UserName, form.Password, form.Referral)

		if err != nil {

			c.JSON(400, gin.H{"status": "error", "message": err.Error()})
			return
		}

		// Generate token if auth is going to perform
		token, firebase := di.generateUserToken(usr.Data().Id)

		// Finished creating the post
		c.JSON(200, gin.H{"status": "okay", "code": 200, "token": token, "firebase": firebase})
		return
	}

	// Couldn't process the request though
	c.JSON(400, gin.H{"status": "error", "message": "Missing information to process the request", "code": 400})
}

func (di *UserAPI) UserGetActivity(c *gin.Context) {

	var activity = make([]model.UserActivity, 0)

	// Get the database interface from the DI
	database := di.DataService.Database
	user_id := c.Param("id")
	kind := c.Param("kind")
	offset := 0
	limit := 10

	if bson.IsObjectIdHex(user_id) == false {

		c.JSON(400, gin.H{"status": "error", "message": "Invalid user id."})
		return
	}

	usr, err := di.User.Get(bson.ObjectIdHex(user_id))

	if err != nil {

		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	query_offset := c.Query("offset")

	if query_offset != "" {

		query_offset_parse, err := strconv.Atoi(query_offset)

		if err == nil {

			offset = query_offset_parse
		}
	}

	query_limit := c.Query("limit")

	if query_limit != "" {

		query_limit_parse, err := strconv.Atoi(query_limit)

		if err == nil {

			limit = query_limit_parse
		}
	}

	switch kind {
	case "comments":

		var commented_posts []model.PostCommentModel
		var commented_count model.PostCommentCountModel

		pipeline_line := []bson.M{
			{
				"$match": bson.M{"users": usr.Data().Id},
			},
			{
				"$unwind": "$comments.set",
			},
			{
				"$project": bson.M{"title": 1, "slug": 1, "comment": "$comments.set"},
			},
			{
				"$match": bson.M{"comment.user_id": usr.Data().Id},
			},
			{
				"$sort": bson.M{"comment.created_at": -1},
			},
		}

		pipeline := database.C("posts").Pipe(append(pipeline_line,
			[]bson.M{
				{
					"$limit": limit,
				},
				{
					"$skip": offset,
				},
			}...,
		))

		err := pipeline.All(&commented_posts)

		if err != nil {
			panic(err)
		}

		pipeline = database.C("posts").Pipe(append(pipeline_line,
			[]bson.M{
				{
					"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}},
				},
			}...,
		))

		err = pipeline.One(&commented_count)

		// No results from the aggregation
		if err != nil {

			commented_count = model.PostCommentCountModel{
				Count: 0,
			}
		}

		for _, post := range commented_posts {

			activity = append(activity, model.UserActivity{
				Id:        post.Id,
				Title:     post.Title,
				Slug:      post.Slug,
				Content:   post.Comment.Content,
				Created:   post.Comment.Created,
				Directive: "commented",
				Author: map[string]string{
					"id":    usr.Data().Id.Hex(),
					"name":  usr.Data().UserName,
					"email": usr.Data().Email,
				},
			})
		}

		// Sort the full set of posts by the time they happened
		sort.Sort(model.ByCreatedAt(activity))

		c.JSON(200, gin.H{"count": commented_count.Count, "activity": activity})

	default:

		c.JSON(400, gin.H{"status": "error", "message": "Invalid request."})
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

func (di *UserAPI) trackSiftScienceLogin(user_id, session_id string, success bool) {

	defer di.Errors.Recover()

	status := "$success"

	if !success {
		status = "$failure"
	}

	err := gosift.Track("$login", map[string]interface{}{
		"$user_id": user_id,
		"$session_id": session_id,
		"$login_status": status,
	})

	if err != nil {
		panic(err)
	}
}
