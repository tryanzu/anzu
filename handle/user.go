package handle

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/dgrijalva/jwt-go"
	u "github.com/fernandez14/spartangeek-blacker/core/user"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/fernandez14/spartangeek-blacker/modules/content"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/security"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
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
	DataService   *mongo.Service               `inject:""`
	CacheService  *goredis.Redis               `inject:""`
	ConfigService *config.Config               `inject:""`
	S3Bucket      *s3.Bucket                   `inject:""`
	User          *user.Module                 `inject:""`
	Content       *content.Module              `inject:""`
	Gaming        *gaming.Module               `inject:""`
	Acl           *acl.Module                  `inject:""`
	Security      *security.Module             `inject:""`
	Collector     CollectorAPI                 `inject:"inline"`
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
		g := di.Gaming.Get(usr)

		g.DailyLogin()
		g.Sync()
	}(usr)

	session_id := c.MustGet("session_id").(string)

	data := usr.Data()
	data.SessionId = session_id

	if len(data.Categories) == 0 {
		data.Categories = make([]bson.ObjectId, 0)
	}

	// Alright, go back and send the user info
	c.JSON(200, data)
}

func (di UserAPI) UserGetJwtToken(c *gin.Context) {
	qs := c.Request.URL.Query()

	// Get the email or the username or the id and its password
	email, password := qs.Get("email"), qs.Get("password")
	usr, err := di.User.Get(bson.M{"email": email, "deleted_at": bson.M{"$exists": false}})

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

	var remember int = 72
	rememberParam := qs.Get("remember")

	if rememberParam != "" {
		rememberTime, rememberErr := strconv.Atoi(rememberParam)

		if rememberErr == nil {
			remember = rememberTime
		}
	}

	// Generate JWT with the information about the user
	token := di.generateUserToken(usr.Data().Id, usr.Data().Roles, remember)

	// Save the activity
	user_id, signed_in := c.Get("user_id")

	permission := c.Query("permission")

	if permission != "" {

		perms := di.Acl.User(usr.Data().Id)

		if perms.Can(permission) == false {
			c.AbortWithStatus(401)
			return
		}
	}

	if signed_in {

		// Save the activity in other routine
		go di.Collector.Activity(model.Activity{UserId: bson.ObjectIdHex(user_id.(string)), Event: "user-view", RelatedId: usr.Data().Id})
	}

	c.JSON(200, gin.H{"status": "okay", "token": token, "session_id": session_id, "expires": remember})
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
			c.JSON(400, gin.H{"status": "error", "message": "Unsupported image type..."})
			return
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
	var form map[string]string

	db := di.DataService.Database
	id := c.MustGet("user_id").(string)
	user_id := bson.ObjectIdHex(id)
	err := db.C("users").FindId(user_id).One(&user)
	if err != nil {
		panic(err)
	}

	if c.BindWith(&form, binding.JSON) == nil {
		set := bson.M{}

		if username, exists := form["username"]; exists && user.NameChanges < 1 {
			valid_username, _ := regexp.Compile(`^[0-9a-zA-Z\-]{0,32}$`)

			if valid_username.MatchString(username) {
				username_slug := sanitize.Path(sanitize.Accents(username))

				// Check whether user exists
				count, _ := db.C("users").Find(bson.M{"username_slug": username_slug}).Count()

				if count == 0 {
					set["username"] = username
					set["username_slug"] = username_slug
					set["name_changes"] = user.NameChanges + 1
				}
			}
		}

		if description, exists := form["description"]; exists {
			if len([]rune(description)) > 60 {
				description = helpers.Truncate(description, 57) + "..."
			}

			set["description"] = description
		}

		if email, exists := form["email"]; exists && user.Email != email {
			if !helpers.IsEmail(email) {
				c.JSON(400, gin.H{"status": "error", "message": "Invalid email address.", "details": "invalid-email", "fields": []string{"email"}})
				return
			}

			_, err := di.User.Get(bson.M{"$or": []bson.M{
				{"email": email},
				{"facebook.email": email},
			}})

			if err == nil {
				c.JSON(400, gin.H{"status": "error", "message": "Email already in use.", "details": "repeated-email", "fields": []string{"email"}})
				return
			}

			set["email"] = email
			set["ver_code"] = helpers.StrRandom(12)
			set["validated"] = false
		}

		if phone, exists := form["phone"]; exists {
			set["phone"] = phone
		}

		if bt, exists := form["battlenet_id"]; exists {
			set["battlenet_id"] = bt
		}

		if steam, exists := form["steam_id"]; exists {
			set["steam_id"] = steam
		}

		if origin, exists := form["origin_id"]; exists {
			set["origin_id"] = origin
		}

		if country, exists := form["country"]; exists && len([]rune(country)) <= 3 {
			set["country"] = country
		}

		if password, exists := form["password"]; exists {

			if len([]rune(password)) < 4 {
				c.JSON(400, gin.H{"status": "error", "message": "Can't allow password update, too short."})
				return
			}

			set["password"] = helpers.Sha256(password)
		}

		set["updated_at"] = time.Now()

		// Update the user profile with some godness
		db.C("users").Update(bson.M{"_id": user.Id}, bson.M{"$set": set})

		if _, emailChanged := set["email"]; emailChanged {
			usr, err := u.FindId(deps.Container, user.Id)
			if err == nil {
				usr.ConfirmationEmail(deps.Container)
			}
		}

		c.JSON(200, gin.H{"status": "okay"})
		return
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

	if c.BindJSON(&form) != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Missing information to process the request", "code": 400})
		return
	}

	usr, err := di.User.SignUp(form.Email, form.UserName, form.Password, form.Referral)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Confirmation email.
	user, err := u.FindId(deps.Container, usr.Data().Id)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	err = user.ConfirmationEmail(deps.Container)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Generate token if auth is going to perform
	token := di.generateUserToken(usr.Data().Id, usr.Data().Roles, 72)

	// Finished creating the post
	c.JSON(200, gin.H{"status": "okay", "code": 200, "token": token})
}

func (di *UserAPI) UserGetActivity(c *gin.Context) {

	var activity = make([]model.UserActivity, 0)

	// Get the database interface from the DI
	database := di.DataService.Database
	content := di.Content
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

		var comments []*feed.Comment
		var posts []struct {
			Id    bson.ObjectId `bson:"_id"`
			Title string        `bson:"title"`
			Slug  string        `bson:"slug"`
		}

		posts_map := make(map[string]struct {
			Id    bson.ObjectId `bson:"_id"`
			Title string        `bson:"title"`
			Slug  string        `bson:"slug"`
		}, 0)

		err := database.C("comments").Find(bson.M{"user_id": usr.Data().Id, "deleted_at": bson.M{"$exists": false}}).Sort("-created_at").Limit(limit).Skip(offset).All(&comments)

		if err != nil {
			panic(err)
		}

		var post_ids []bson.ObjectId

		for _, c := range comments {
			post_ids = append(post_ids, c.PostId)
		}

		err = database.C("posts").Find(bson.M{"_id": bson.M{"$in": post_ids}}).Select(bson.M{"title": 1, "slug": 1}).All(&posts)

		if err != nil {
			panic(err)
		}

		for _, p := range posts {
			posts_map[p.Id.Hex()] = p
		}

		count, err := database.C("comments").Find(bson.M{"user_id": usr.Data().Id, "deleted_at": bson.M{"$exists": false}}).Count()

		if err != nil {
			panic(err)
		}

		for _, c := range comments {

			post, exists := posts_map[c.PostId.Hex()]

			if exists {

				content.ParseTags(c)
				activity = append(activity, model.UserActivity{
					Id:        post.Id,
					Title:     post.Title,
					Slug:      post.Slug,
					Content:   c.Content,
					Created:   c.Created,
					Directive: "commented",
					Author: map[string]string{
						"id":   usr.Data().Id.Hex(),
						"name": usr.Data().UserName,
					},
				})
			}
		}

		// Sort the full set of posts by the time they happened
		sort.Sort(model.ByCreatedAt(activity))

		c.JSON(200, gin.H{"count": count, "activity": activity})

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

type UserToken struct {
	UserID string   `json:"user_id"`
	Scopes []string `json:"scope"`
	jwt.StandardClaims
}

func (di *UserAPI) generateUserToken(id bson.ObjectId, roles []user.UserRole, expiration int) string {

	scope := []string{}
	for _, role := range roles {
		scope = append(scope, role.Name)
	}

	claims := UserToken{
		id.Hex(),
		scope,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(expiration)).Unix(),
			Issuer:    "spartangeek",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Use the secret inside the configuration to encrypt it
	secret, err := di.ConfigService.String("application.secret")
	if err != nil {
		panic(err)
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}

	return tokenString
}
