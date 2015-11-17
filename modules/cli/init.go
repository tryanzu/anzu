package cli

import (
	"encoding/json"
	"fmt"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/modules/search"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"gopkg.in/mgo.v2/bson"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Module struct {
	Mongo   *mongo.Service               `inject:""`
	Search  *search.Module               `inject:""`
	Errors  *exceptions.ExceptionsModule `inject:""`
	User    *user.Module                 `inject:""`
	Feed    *feed.FeedModule             `inject:""`
}

type fn func()

func (module Module) Run(name string) {

	commands := map[string]fn{
		"slug-fix":    module.SlugFix,
		"codes-fix":   module.Codes,
		"index-posts": module.IndexAlgolia,
		"index-components": module.IndexComponentsAlgolia,
		"send-confirmations": module.ConfirmationEmails,	
		"replace-url": module.ReplaceURL,
	}

	if handler, exists := commands[name]; exists {

		handler()
		return
	}

	// If reachs this point then panic
	log.Panic("No such handler for cli")
}

func (module Module) ReplaceURL() {

	var usr user.User
	var post model.Post

	database := module.Mongo.Database

	// Get all users
	iter := database.C("users").Find(nil).Iter()

	for iter.Next(&usr) {

		image := strings.Replace(usr.Image, "http://s3-us-west-1.amazonaws.com/spartan-board", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
		image  = strings.Replace(usr.Image, "http://assets.spartangeek.com", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
		image  = strings.Replace(usr.Image, "https://assets.spartangeek.com", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
		err := database.C("users").Update(bson.M{"_id": usr.Id}, bson.M{"$set": bson.M{"image": image}})

		if err != nil {
			continue
		}

		fmt.Printf(".")
	}

	// Get all posts
	iter = database.C("posts").Find(nil).Iter()

	for iter.Next(&post) {

		updates := bson.M{}

		content := strings.Replace(post.Content, "https://assets.spartangeek.com", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
		content  = strings.Replace(post.Content, "http://assets.spartangeek.com", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
		content  = strings.Replace(post.Content, "http://s3-us-west-1.amazonaws.com/spartan-board", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
		updates["content"] = content
		

		for index, comment := range post.Comments.Set {

			comment_index := strconv.Itoa(index)

			content := strings.Replace(comment.Content, "https://assets.spartangeek.com", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
			content  = strings.Replace(comment.Content, "http://assets.spartangeek.com", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
			content  = strings.Replace(comment.Content, "http://s3-us-west-1.amazonaws.com/spartan-board", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
		

			updates["comments.set." + comment_index + ".content"] = content
		}

		if len(updates) > 0 {

			err := database.C("posts").Update(bson.M{"_id": post.Id}, bson.M{"$set": updates})

			if err != nil {
				continue
			}

			fmt.Printf("$")
		}
	}
}

func (module Module) SlugFix() {

	var usr user.User
	database := module.Mongo.Database
	valid_name, _ := regexp.Compile(`^[a-zA-Z][a-zA-Z0-9]*[._-]?[a-zA-Z0-9]+$`)

	// Get all users
	iter := database.C("users").Find(nil).Select(bson.M{"_id": 1, "username": 1, "email": 1, "username_slug": 1}).Iter()

	for iter.Next(&usr) {

		slug := helpers.StrSlug(usr.UserName)

		if !valid_name.MatchString(usr.UserName) {

			// Fallback username to slug
			err := database.C("users").Update(bson.M{"_id": usr.Id}, bson.M{"$set": bson.M{"username": slug, "username_slug": slug, "name_changes": 0}})

			if err != nil {
				panic(err)
			}

			log.Printf("\n%v --> %v\n", usr.UserName, slug)
			continue

		} else {

			// Fix slug in case they need if
			if slug != usr.UserNameSlug {

				err := database.C("users").Update(bson.M{"_id": usr.Id}, bson.M{"$set": bson.M{"username_slug": slug}})

				if err != nil {
					panic(err)
				}

				fmt.Printf("-")
				log.Printf("\n%v -slug-> %v\n", usr.UserNameSlug, slug)
				continue
			}
		}

		fmt.Printf(".")
	}
}

func (module Module) Codes() {

	var usr user.User
	database := module.Mongo.Database

	// Get all users
	iter := database.C("users").Find(nil).Select(bson.M{"_id": 1, "ref_code": 1, "ver_code": 1}).Iter()

	for iter.Next(&usr) {

		if usr.VerificationCode == "" {

			code := helpers.StrRandom(12)
			err := database.C("users").Update(bson.M{"_id": usr.Id}, bson.M{"$set": bson.M{"ver_code": code, "validated": false}})

			if err != nil {
				panic(err)
			}

			fmt.Printf("+")
		}

		if usr.ReferralCode == "" {

			code := helpers.StrRandom(6)
			err := database.C("users").Update(bson.M{"_id": usr.Id}, bson.M{"$set": bson.M{"ref_code": code}})

			if err != nil {
				panic(err)
			}

			fmt.Printf("-")
		}

		if usr.ReferralCode != "" && usr.VerificationCode != "" {

			fmt.Printf(".")
		}
	}
}

func (module Module) ConfirmationEmails() {

	var usr user.User
	database := module.Mongo.Database

	// Get all users
	iter := database.C("users").Find(bson.M{"validated": false}).Iter()

	for iter.Next(&usr) {

		usr_copy := &usr
		usr_obj, err := module.User.Get(usr_copy)
			
		if err == nil {

			// Send the confirmation email
			usr_obj.SendConfirmationEmail()
		}
	}
}

func (module Module) IndexAlgolia() {

	var post model.Post
	var categories []model.Category

	database := module.Mongo.Database
	index := module.Search.Get("board")
	ocategories := make(map[string]model.Category)

	// Get categories
	err := database.C("categories").Find(nil).All(&categories)

	if err != nil {
		panic(err)
	}

	for _, category := range categories {

		ocategories[category.Id.Hex()] = category
	}

	// Get an iterator to work with the posts in an efficient way
	iter := database.C("posts").Find(bson.M{"deleted": bson.M{"$exists": false}}).Sort("-pinned", "-created_at").Iter()

	// Prepare batch variables
	batch_count := 0
	batch_store := make([]feed.AlgoliaPostModel, 0)

	for iter.Next(&post) {

		// Unefficent in some ways but DRY
		user, err := module.User.Get(post.UserId)

		if err != nil {

			fmt.Printf("&")
			continue
		}

		user_data := user.Data()
		tribute := post.Votes.Up
		shit := post.Votes.Down

		for _, comment := range post.Comments.Set {

			tribute = tribute + comment.Votes.Up
			shit = shit + comment.Votes.Down
		}

		if post_category, exists := ocategories[post.Category.Hex()]; exists {

			final_rate := 0.0

			// Calculate feed rate at given time
			post_obj, err := module.Feed.Post(post)

			if err != nil {

				fmt.Printf("$")
				continue
			}

			reached, viewed := post_obj.GetReachViews(post.Id)
			total := reached + viewed

			if total > 101 {

				if reached == 0 {

					reached = 1
				}

				if viewed == 0 {

					viewed = 1
				}

				view_rate := 100.0 / float64(reached) * float64(viewed)
				comment_rate := 100.0 / float64(viewed) * float64(post.Comments.Count)
				final_rate = (view_rate + comment_rate) / 2.0
			}

			components := make([]string, 0)

			// If the post is a recommendations post then reflect to get the components
			if post.Type == "recommendations" {

				bindable := reflect.ValueOf(&post.Components).Elem()

				for i := 0; i < bindable.NumField(); i++ {

					field := bindable.Field(i).Interface()

					switch field.(type) {
					case model.Component:

						component := field.(model.Component)

						if component.Content != "" {

							components = append(components, component.Content)
						}

					default:
						continue
					}
				}
			}

			item := feed.AlgoliaPostModel{
				Id:       post.Id.Hex(),
				Title:    post.Title,
				Content:  post.Content,
				Comments: post.Comments.Count,
				Slug:     post.Slug,
				User: feed.AlgoliaUserModel{
					Id:       user_data.Id.Hex(),
					Username: user_data.UserName,
					Image:    user_data.Image,
					Email:    user_data.Email,
				},
				Category: feed.AlgoliaCategoryModel{
					Id:   post.Category.Hex(),
					Name: post_category.Name,
				},
				Popularity: final_rate,
				Created:    post.Created.Unix(),
				Components: components,
			}

			fmt.Printf("+")

			// Append to the current batch
			batch_count++
			batch_store = append(batch_store, item)

			if batch_count >= 1000 {

				var json_objects []interface{}
				json_data, err := json.Marshal(batch_store)

				if err != nil {

					fmt.Printf("\n%v\n", batch_store)

					panic(err)
				}

				err = json.Unmarshal(json_data, &json_objects)

				if err != nil {
					panic(err)
				}

				_, err = index.UpdateObjects(json_objects)

				if err != nil {
					panic(err)
				}

				batch_store = make([]feed.AlgoliaPostModel, 0)
				batch_count = 0
			}

			continue
		}

		// Process the last incomplete batch
		var json_objects []interface{}
		json_data, err := json.Marshal(batch_store)

		if err != nil {

			fmt.Printf("\n%v\n", batch_store)

			panic(err)
		}

		err = json.Unmarshal(json_data, &json_objects)

		if err != nil {
			panic(err)
		}

		_, err = index.UpdateObjects(json_objects)

		if err != nil {
			panic(err)
		}

		fmt.Printf("*")
	}
}

func (module Module) IndexComponentsAlgolia() {

	var component map[string]interface{}

	database := module.Mongo.Database
	index := module.Search.Get("components")

	// Get an iterator to work with the posts in an efficient way
	iter := database.C("components").Find(nil).Iter()

	// Prepare batch variables
	batch_count := 0
	batch_store := make([]components.AlgoliaComponentModel, 0)
	batch_check := func(last_check bool) {

		if last_check || batch_count >= 1000 {

			var json_objects []interface{}
			json_data, err := json.Marshal(batch_store)

			if err != nil {
				panic(err)
			}

			err = json.Unmarshal(json_data, &json_objects)

			if err != nil {
				panic(err)
			}

			_, err = index.UpdateObjects(json_objects)

			if err != nil {
				panic(err)
			}

			batch_store = make([]components.AlgoliaComponentModel, 0)
			batch_count = 0
		}
	}

	for iter.Next(&component) {

		if name, name_exists := component["name"]; name_exists {

			full_name, full_name_exists := component["full_name"]

			if ! full_name_exists {

				full_name = name
			}

			images := component["images"].([]map[string]string)
			image := ""

			if len(images) > 0 {

				image = images[0]["path"]
				image = strings.Replace(image, "full/", "", -1)

			} else {

				image = ""
			}

			id := component["_id"].(bson.ObjectId)
			part_number := component["part_number"].(string)
			slug := component["slug"].(string)

			item := components.AlgoliaComponentModel{
				Id: id.Hex(),
				Name: name.(string),
				FullName: full_name.(string),
				Part: part_number,
				Slug: slug,
				Image: image,
			}

			fmt.Printf("+")

			// Append to the current batch
			batch_count++
			batch_store = append(batch_store, item)

			batch_check(false)
		}
	}

	// Upload last batch even if incomplete
	batch_check(true)

	fmt.Printf("*")
}
