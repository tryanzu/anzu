package cli

import (
	"github.com/algolia/algoliasearch-client-go/algoliasearch"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/fernandez14/spartangeek-blacker/model"
	"gopkg.in/mgo.v2/bson"
	"log"
	"fmt"
	"regexp"
	"encoding/json"
)

type Module struct {
	Mongo   *mongo.Service               `inject:""`
	Algolia *algoliasearch.Index         `inject:""`
	Errors  *exceptions.ExceptionsModule `inject:""`
	User    *user.Module                 `inject:""`
	Feed    *feed.FeedModule             `inject:""`
}

type fn func()

func (module Module) Run(name string) {

	commands := map[string]fn {
		"slug-fix": module.SlugFix,
		"codes-fix": module.Codes,
		"index-posts": module.IndexAlgolia,
	}

	if handler, exists := commands[name]; exists {

		handler()
		return
	} 

	// If reachs this point then panic
	log.Panic("No such handler for cli")
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

func (module Module) IndexAlgolia() {

	var post model.Post
	var categories []model.Category

	database := module.Mongo.Database
	index := module.Algolia
	feed := module.Feed
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
	iter := database.C("posts").Find(nil).Iter()

	// Prepare batch variables
	batch_count := 0
	batch_store := make([]PostModel, 1000)

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
			post_obj, err := feed.Post(post)

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

			item := PostModel{
				Id: post.Id.Hex(),
				Title: post.Title,
				Content: post.Content,
				Comments: post.Comments.Count,
				User: UserModel{
					Id: user_data.Id.Hex(),
					Username: user_data.UserName,
					Image: user_data.Image,
					Email: user_data.Email,
				},
				Category: CategoryModel{
					Id: post.Category.Hex(),
					Name: post_category.Name,
				},
				Popularity: final_rate,
				Created: post.Created.Unix(),
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

				_, err = index.AddObjects(json_objects)

				if err != nil {
					panic(err)
				}			

				batch_store = make([]PostModel, 1000)
				batch_count = 0
			}

			continue
		}

		fmt.Printf("*")
	}	
}