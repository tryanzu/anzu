package main

import (
    "net/http"
    "crypto/sha256"
    "encoding/hex"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "github.com/martini-contrib/render"
    "code.google.com/p/go-uuid/uuid"
    "time"
)

type User struct {
    Id bson.ObjectId `bson:"_id,omitempty" json:"id"`
    FirstName    string `bson:"first_name" json:"first_name"`
    LastName     string `bson:"last_name" json:"last_name"`
    UserName     string `bson:"username" json:"username"`
    Password     string `bson:"password" json:"-"`
    Email        string `bson:"email" json:"email"`
    Roles        []string `bson:"roles" json:"roles"`
    Permissions  []string `bson:"permissions" json:"permissions"`
    Description  string `bson:"description" json:"description"`
    Created      time.Time `bson:"created_at" json:"created_at"`
    Updated      time.Time `bson:"updated_at" json:"updated_at"`
}

type UserToken struct {
    Id      bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
    UserId  bson.ObjectId `bson:"user_id" json:"user_id"`
    Token   string `bson:"token" json:"token"`
    Created time.Time `bson:"created_at" json:"created_at"`
    Updated time.Time `bson:"updated_at" json:"updated_at"`
}

func UserGetOne (r render.Render, database *mgo.Database, req *http.Request) {
    
   
    
}

func UserGetByToken (r render.Render, database *mgo.Database, req *http.Request) {
    
    // Users collection    
    collection := database.C("users")
    
    // Get user by token
    user_token  := UserToken{}
    token := req.Header.Get("Auth-Token")
    
    // Try to fetch the user using token header though
	err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)
	
	if err != nil {
	 
	    response := map[string]string{
		    "error":  "Couldnt found user using that token.",
		    "status": "103",
	    }
        
        r.JSON(404, response)
        
        return   
	}
	
	
	user := User{}
	
	err = collection.Find(bson.M{"_id": user_token.UserId}).One(&user)
	
	if err != nil {
	    
	    response := map[string]string{
		    "error":  "Couldnt found user using that token.",
		    "status": "104",
	    }
        
        r.JSON(404, response)
        
        return  
	}
	
	// Alright, go back and send the user info
	r.JSON(200, user)
}

func UserGetToken (r render.Render, database *mgo.Database, req *http.Request) {
    
    // Get the query parameters
    qs := req.URL.Query()
    
    // Get the email or the username or the id and its password
    email, password := qs.Get("email"), qs.Get("password")
    
    collection := database.C("users")
    
    user := User{}
    
    // Try to fetch the user using email param though
	err := collection.Find(bson.M{"email": email}).One(&user)
 
	if err != nil {
	    
	    response := map[string]string{
		    "error":  "Couldnt found user with that email.",
		    "status": "101",
	    }
        
        r.JSON(404, response)
        
        return
	}
    
    // Incorrect password
    password_encrypted := []byte(password)
    sha256 := sha256.New()
    sha256.Write(password_encrypted)
    md := sha256.Sum(nil)
    hash := hex.EncodeToString(md)
    
    if user.Password != hash {
        response := map[string]string{
		    "error":  "The keys used for authentication are not correct.",
		    "status": "102",
	    }
        
        r.JSON(404, response)
        
        return
    }  
    
    // Generate user token
    uuid := uuid.New()
    token := &UserToken {
        UserId: user.Id, 
        Token: uuid, 
        Created: time.Now(), 
        Updated: time.Now(),
    }
    
    err = database.C("tokens").Insert(token)
    
    r.JSON(200, token)    
}