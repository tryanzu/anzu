package main

import (
    "net/http"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"
    "time"
    "reflect"
    "encoding/json"
    "io/ioutil"
)

type Votes struct {
    Up string `bson:"up" json:"up"`
    Down string `bson:"down" json:"down"`
    Rating string `bson:"rating,omitempty" json:"rating,omitempty"`
}

type Author struct {
    Id     bson.ObjectId `bson:"id,omitempty" json:"id,omitempty"`
    Name   string `bson:"name" json:"name"`
    Email  string `bson:"email" json:"email"`
    Avatar string `bson:"avatar" json:"avatar"`
    Profile interface{} `bson:"profile,omitempty" json:"profile,omitempty"`
}

type Comments struct {
    Count int `bson:"count" json:"count"` 
    Set   []Comment `bson:"set" json:"set"`
}

type Comment struct {
    UserId  bson.ObjectId `bson:"user_id" json:"user_id"`
    Votes   Votes `bson:"votes" json:"votes"`
    User    interface{} `bson:"author,omitempty" json:"author,omitempty"`
    Content string `bson:"content" json:"content"`
    Created time.Time `bson:"created_at" json:"created_at"`
}

type Components struct {
    Cpu Component `bson:"cpu,omitempty" json:"cpu,omitempty"` 
    Motherboard Component `bson:"motherboard,omitempty" json:"motherboard,omitempty"` 
    Ram Component `bson:"ram,omitempty" json:"ram,omitempty"` 
    Storage Component `bson:"storage,omitempty" json:"storage,omitempty"` 
    Cooler Component `bson:"cooler,omitempty" json:"cooler,omitempty"` 
    Power Component `bson:"power,omitempty" json:"power,omitempty"` 
    Cabinet Component `bson:"cabinet,omitempty" json:"cabinet,omitempty"` 
    Screen Component `bson:"screen,omitempty" json:"screen,omitempty"` 
    Videocard Component `bson:"videocard,omitempty" json:"videocard,omitempty"` 
    Software string `bson:"software,omitempty" json:"software,omitempty"` 
    Budget   string `bson:"budget,omitempty" json:"budget,omitempty"` 
    BudgetCurrency string `bson:"budget_currency,omitempty" json:"budget_currency,omitempty"` 
    BudgetType string `bson:"budget_type,omitempty" json:"budget_type,omitempty"` 
    BudgetFlexibility string `bson:"budget_flexibility,omitempty" json:"budget_flexibility,omitempty"` 
}

type Component struct {
    Content   string `bson:"content" json:"content"` 
    Elections bool   `bson:"elections" json:"elections"` 
    Options   []ElectionOption `bson:"options,omitempty" json:"options,omitempty"` 
    Votes     Votes  `bson:"votes" json:"votes"`
    Status    string `bson:"status" json:"status"` 
}

type ElectionOption struct {
    UserId  bson.ObjectId `bson:"user_id" json:"user_id"`
    Content string `bson:"content" json:"content"`
    User    interface{} `bson:"author,omitempty" json:"author,omitempty"`
    Votes   Votes `bson:"votes" json:"votes"`
    Created time.Time `bson:"created_at" json:"created_at"`
}

type Post struct {
    Id bson.ObjectId `bson:"_id" json:"id"`
    Title string `bson:"title" json:"title"`   
    Slug  string `bson:"slug" json:"slug"`   
    Type  string `bson:"type" json:"type"`   
    Content string `bson:"content" json:"content"`   
    Categories []string `bson:"categories" json:"categories"`   
    Comments   Comments `bson:"comments" json:"comments"`   
    Author     User `bson:"author" json:"author"`  
    UserId     bson.ObjectId `bson:"user_id,omitempty" json:"user_id,omitempty"`
    Users      []bson.ObjectId `bson:"users,omitempty" json:"users,omitempty"`
    Votes      Votes `bson:"votes" json:"votes"`
    Components Components `bson:"components" json:"components"`
    Created time.Time `bson:"created_at" json:"created_at"`
    Updated time.Time `bson:"updated_at" json:"updated_at"`
}

func PostsGet (r render.Render, database *mgo.Database, req *http.Request) {
    
    // Get the query parameters
    qs := req.URL.Query()
    
    // Name of the set to get
    named := qs.Get("named")
    
    var results []Post
    
    // Get the collection
    collection := database.C("posts")
    query := collection.Find(bson.M{}).Limit(10)
    
    if named == "" || named == "default" {
        
        // Sort from newest to oldest
        query = query.Sort("-created_at")
    }
    
    
    if named == "most-loved" {
        
        // Sort from the most loved to the no loved =(
        query = query.Sort("-votes.rating")
    }
    
    if named == "most-commented" {
        
        // Sort from the most loved to the no loved =(
        query = query.Sort("-comments.count")
    }
    
    // Try to fetch the posts
	err := query.All(&results)
	
	if err != nil {
	    
	    panic(err)
	}
    
    var authors []bson.ObjectId
    
    for _, post := range results {
        
        authors = append(authors, post.UserId)
    }
    
    var users []User
        
    // Get the users
    collection = database.C("users")
    
    err = collection.Find(bson.M{"_id": bson.M{"$in": authors}}).All(&users)
    
    if err != nil {
        
        panic(err)   
    }
    
    usersMap := make(map[bson.ObjectId]User)
    
    for _, user := range users {
        
        usersMap[user.Id] = user
    }
    
    for index := range results  {
        
        post := &results[index]
        
        if _, okay := usersMap[post.UserId]; okay {

            postUser := usersMap[post.UserId]
            
            post.Author = User {
                Id: postUser.Id,
                UserName: postUser.UserName,
                FirstName: postUser.FirstName,
                LastName: postUser.LastName,
            }
        } 
    }
    
    r.JSON(200, results)
}

func PostsGetOne (r render.Render, database *mgo.Database, req *http.Request, params martini.Params) {
    
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

func PostsGetOneSlug (r render.Render, database *mgo.Database, req *http.Request, params martini.Params) {
    
    // Get the post using the slug
    slug := params["slug"]
    
    // Get the collection
    collection := database.C("posts")
    
    post := Post{}
    
    // Try to fetch the needed post by id
    err := collection.Find(bson.M{"slug": slug}).One(&post)
    
    if err != nil {
        
        response := map[string]string{
		    "error":  "Couldnt found post with that slug.",
		    "status": "203",
	    }
        
        r.JSON(404, response)
        
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
                "id": user.Id.Hex(),
                "name": user.UserName,
                "email": user.Email,
            }
        }
        
        for index := range post.Comments.Set  {
            
            comment := &post.Comments.Set[index]
            
            if _, okay := usersMap[comment.UserId]; okay {
                
                post.Comments.Set[index].User = usersMap[comment.UserId]   
            } 
        }
        
        components := reflect.ValueOf(&post.Components).Elem()
        
        for i := 0; i < components.NumField(); i++ {
            
            f := components.Field(i)
            
            if f.Type().String() == "main.Component" {
                
                component := f.Interface().(Component)
                
                if component.Elections == true {
                    
                    for option_index, option := range component.Options {
                        
                        if _, okay := usersMap[option.UserId]; okay {
                            
                            component.Options[option_index].User = usersMap[option.UserId]
                        }    
                    }
                    
                    f.Set(reflect.ValueOf(component))
                }
            }
        }
    }
    
    r.JSON(200, post)  
}

func PostCreate (r render.Render, database *mgo.Database, req *http.Request) {
    
    body, err := ioutil.ReadAll(req.Body)
    
    if err != nil {
        
        panic(err)   
    }
    
    var post map[string] interface{}
    
    err = json.Unmarshal(body, &post)
    
    if err != nil {
        
        panic(err)   
    }
    
    kind, kind_okay := post["kind"]
    
    if kind_okay && (kind == "recommendations" || kind == "") {
        
    
        // Validate the form fields, before to create the post we need to validate the components
        _, name_present         := post["name"]
        _, content_present      := post["content"]
        _, budget_present       := post["budget"]
        _, currency_present     := post["currency"]
        _, flexibility_present  := post["moves"]
        _, software_present     := post["software"]
        _, category_present     := post["tag"]
        
        // Check the presence of every single needed param
        if ! name_present || ! content_present || ! budget_present || ! currency_present || ! flexibility_present || ! software_present || ! category_present {
            
            response := map[string]string{
        	    "error":  "Couldnt create the post due to malformations.",
        	    "status": "204",
            }
            
            r.JSON(400, response)
            
            return
        }
        
        r.JSON(200, post)
        
        return
    }
    
    response := map[string]string{
	    "error":  "Couldnt create the post due to malformations.",
	    "status": "205",
    }
    
    r.JSON(400, response)

}