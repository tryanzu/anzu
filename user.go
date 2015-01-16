package main

import (
	"code.google.com/p/go-uuid/uuid"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/mrvdot/golang-utils"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type User struct {
	Id          bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	FirstName   string                 `bson:"first_name" json:"first_name"`
	LastName    string                 `bson:"last_name" json:"last_name"`
	UserName    string                 `bson:"username" json:"username"`
	Password    string                 `bson:"password" json:"-"`
	Email       string                 `bson:"email" json:"email,omitempty"`
	Roles       []string               `bson:"roles" json:"roles,omitempty"`
	Permissions []string               `bson:"permissions" json:"permissions,omitempty"`
	Description string                 `bson:"description" json:"description,omitempty"`
	Facebook    interface{}            `bson:"facebook,omitempty" json:"facebook,omitempty"`
	Profile     map[string]interface{} `bson:"profile,omitempty" json:"profile,omitempty"`
	Stats       UserStats              `bson:"stats,omitempty" json:"stats,omitempty"`
	Created     time.Time              `bson:"created_at" json:"created_at"`
	Updated     time.Time              `bson:"updated_at" json:"updated_at"`
}

type UserStats struct {
	Saw int `bson:"saw" json:"saw"`
}

type UserToken struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	UserId  bson.ObjectId `bson:"user_id" json:"user_id"`
	Token   string        `bson:"token" json:"token"`
	Closed  bool          `bson:"closed,omitempty" json"closed,omitempty"`
	Created time.Time     `bson:"created_at" json:"created_at"`
	Updated time.Time     `bson:"updated_at" json:"updated_at"`
}

type UserFollowing struct {
	Id            bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Follower      bson.ObjectId `bson:"follower,omitempty" json:"follower"`
	Following     bson.ObjectId `bson:"following,omitempty" json:"following"`
	Notifications bool          `bson:"notifications,omitempty" json:"notifications"`
	Created       time.Time     `bson:"created_at" json:"created_at"`
}

type UserProfileForm struct {
	Country       string `json:"country" binding:"required"`
	FavouriteGame string `json:"favourite_game" binding:"required"`
	Microsoft     string `json:"microsoft" binding:"required"`
	Biography     string `json:"bio" binding:"required"`
	ShowEmail     bool   `json:"show_email" binding:"required"`
}

type UserRegisterForm struct {
	UserName  string  `json:"username" binding:"required"`
	Password  string  `json:"password" binding:"required"`
	Email     string  `json:"email" binding:"required"`
}

func UserGetByToken(c *gin.Context) {

	// Users collection
	collection := database.C("users")

	// Get user by token
	user_token := UserToken{}
	token := c.Request.Header.Get("Auth-Token")

	// Try to fetch the user using token header though
	err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)

	if err == nil {

		user := User{}
		err = collection.Find(bson.M{"_id": user_token.UserId}).One(&user)

		if err == nil {

			// Alright, go back and send the user info
			c.JSON(200, user)

			return
		}
	}

	c.JSON(401, gin.H{"error": "Could find user by token...", "status": 103})
}

func UserGetToken(c *gin.Context) {

	// Get the query parameters
	qs := c.Request.URL.Query()

	// Get the email or the username or the id and its password
	email, password := qs.Get("email"), qs.Get("password")

	collection := database.C("users")

	user := User{}

	// Try to fetch the user using email param though
	err := collection.Find(bson.M{"email": email}).One(&user)

	if err != nil {

		c.JSON(401, gin.H{"error": "Couldnt found user with that email", "status": 101})
		return
	}

	// Incorrect password
	password_encrypted := []byte(password)
	sha256 := sha256.New()
	sha256.Write(password_encrypted)
	md := sha256.Sum(nil)
	hash := hex.EncodeToString(md)

	if user.Password != hash {

		c.JSON(401, gin.H{"error": "Credentials are not correct.", "status": 102})
		return
	}

	// Generate user token
	uuid := uuid.New()
	token := &UserToken{
		UserId:  user.Id,
		Token:   uuid,
		Closed:  false,
		Created: time.Now(),
		Updated: time.Now(),
	}

	err = database.C("tokens").Insert(token)

	c.JSON(200, token)
}

func UserGetTokenFacebook(c *gin.Context) {

	var facebook map[string]interface{}

	// Bind to map
	c.BindWith(&facebook, binding.JSON)

	facebook_id := facebook["id"]

	// Validate the facebook id
	if facebook_id == "" {

		c.JSON(401, gin.H{"error": "Invalid oAuth facebook token...", "status": 105})
		return
	}

	collection := database.C("users")
	user := User{}

	// Try to fetch the user using the facebook id param though
	err := collection.Find(bson.M{"facebook.id": facebook_id}).One(&user)

	// Create a new user
	if err != nil {

		username := facebook["first_name"].(string) + " " + facebook["last_name"].(string)
		id := bson.NewObjectId()

		user := &User{
			Id:          id,
			FirstName:   facebook["first_name"].(string),
			LastName:    facebook["last_name"].(string),
			UserName:    utils.GenerateSlug(username),
			Password:    "",
			Email:       facebook["email"].(string),
			Roles:       make([]string, 0),
			Permissions: make([]string, 0),
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

		// Generate user token
		uuid := uuid.New()
		token := &UserToken{
			UserId:  id,
			Token:   uuid,
			Created: time.Now(),
			Updated: time.Now(),
		}

		err = database.C("tokens").Insert(token)

		if err != nil {

			panic(err)
		}

		c.JSON(200, token)

	} else {

		// Generate user token
		uuid := uuid.New()
		token := &UserToken{
			UserId:  user.Id,
			Token:   uuid,
			Closed:  false,
			Created: time.Now(),
			Updated: time.Now(),
		}

		err = database.C("tokens").Insert(token)

		if err != nil {

			panic(err)
		}

		c.JSON(200, token)
	}
}

func UserUpdateProfile(c *gin.Context) {

	// Users collection
	users_collection := database.C("users")

	// Get user by token
	user_token := UserToken{}
	token := c.Request.Header.Get("Auth-Token")

	// Try to fetch the user using token header though
	err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)

	if err == nil {

		user := User{}
		err = users_collection.Find(bson.M{"_id": user_token.UserId}).One(&user)

		if err == nil {

			var profileUpdate UserProfileForm

			if c.BindWith(&profileUpdate, binding.JSON) {

				changes := bson.M{"$set": bson.M{
					"profile.country":        profileUpdate.Country,
					"profile.favourite_game": profileUpdate.FavouriteGame,
					"profile.microsoft":      profileUpdate.Microsoft,
					"profile.bio":            profileUpdate.Biography,
					"updated_at":             time.Now(),
				}}

				// Update the user profile with some godness
				users_collection.Update(bson.M{"_id": user.Id}, changes)

				c.JSON(200, gin.H{"message": "okay", "status": 900})
				return
			}
		}
	}
}

func UserRegisterAction(c *gin.Context) {

    var registerAction UserRegisterForm

    if c.BindWith(&registerAction, binding.JSON) {

        // Check if already registered
        email_exists, _ := database.C("users").Find(bson.M{"email": registerAction.Email}).Count()

        if email_exists > 0 {

            // Only one account per email
            c.JSON(400, gin.H{"status": "error", "message": "User already registered", "code": 470})
            return
        }

        user_exists, _ := database.C("users").Find(bson.M{"username": registerAction.UserName}).Count()

        if user_exists > 0 {

            // Username busy
            c.JSON(400, gin.H{"status": "error", "message": "User already registered", "code": 470})
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
	        "step": 0,
	        "ranking": 0,
	        "country": "México",
	        "posts": 0,
	        "followers": 0,
	        "show_email": false,
	        "favourite_game": "-",
	        "microsoft": "-",
	        "bio": "Just another spartan geek",
	    }

        user := &User{
            FirstName: "",
            LastName: "",
            UserName: registerAction.UserName,
            Password: hash,
            Email: registerAction.Email,
            Roles: []string{"registered"},
            Permissions: make([]string, 0),
            Description: "",
            Profile: profile,
            Stats: UserStats{
                Saw: 0,
            },
            Created: time.Now(),
            Updated: time.Now(),
        }
        
        err := database.C("users").Insert(user)

        if err != nil {
            panic(err)
        }
        
        // Send a confirmation email
        

        // Finished creating the post
		c.JSON(200, gin.H{"status": "okay", "code": 200})
		return
    }
    
    // Couldn't process the request though
    c.JSON(400, gin.H{"status": "error", "message": "Missing information to process the request", "code": 400})
}