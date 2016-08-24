package cli

import (
	"encoding/json"
	"fmt"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/search"
	"github.com/fernandez14/spartangeek-blacker/modules/transmit"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/olebedev/config"
	"gopkg.in/jmcvetta/neoism.v1"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/op/go-logging.v1"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Module struct {
	Mongo     *mongo.Service               `inject:""`
	Search    *search.Module               `inject:""`
	Errors    *exceptions.ExceptionsModule `inject:""`
	User      *user.Module                 `inject:""`
	Feed      *feed.FeedModule             `inject:""`
	Transmit  *transmit.Sender             `inject:""`
	Mail      *mail.Module                 `inject:""`
	GCommerce *gcommerce.Module            `inject:""`
	Config    *config.Config               `inject:""`
	Neoism    *neoism.Database             `inject:""`
	Logger    *logging.Logger              `inject:""`
}

type fn func()

func (module Module) Run(name string) {

	commands := map[string]fn{
		"slug-fix":           module.SlugFix,
		"codes-fix":          module.Codes,
		"index-posts":        module.IndexAlgolia,
		"index-components":   module.IndexComponentsAlgolia,
		"send-confirmations": module.ConfirmationEmails,
		"replace-url":        module.ReplaceURL,
		"test-transmit":      module.TestSocket,
		"first-newsletter":   module.FirstNewsletter,
		"massdrop-invoicing": module.GenerateMassdropInvoices,
		"custom-invoicing":   module.GenerateCustomInvoice,
		"custom-invoices":    module.GenerateCustomInvoices,
		"migrate-comments":   module.MigrateDeletedComment,
		"migrate-ccomments":  module.MigrateChosenComment,
		"export-components":  module.ExportComponents,
		"count-components":   module.GenerateComponentViews,

		"spree-products":            module.SpreeProducts,
		"spree-products-images":     module.SpreeProductsImages,
		"spree-products-flush":      module.SpreeProductsFlush,
		"spree-products-properties": module.SpreeProductsProperties,
	}

	if handler, exists := commands[name]; exists {

		handler()
		return
	}

	// If reachs this point then panic
	log.Panic("No such handler for cli")
}

func (module Module) FirstNewsletter() {

	var news NewsletterModel
	var usr user.UserPrivate

	database := module.Mongo.Database
	iter := database.C("newsletter").Find(nil).Limit(200000).Iter()

	for iter.Next(&news) {

		email := news.Value.Email
		err := database.C("users").Find(bson.M{"$or": []bson.M{{"email": email}, {"facebook.email": email}}}).One(&usr)

		if err == nil {

			gamificated := usr.Gamificated
			since := time.Since(gamificated)

			if since.Hours() < float64(60.0*24.0*20) {
				database.C("newsletter").Update(bson.M{"_id": news.Id}, bson.M{"$set": bson.M{"old": false}})
				fmt.Printf("-")
				continue
			}
		}

		database.C("newsletter").Update(bson.M{"_id": news.Id}, bson.M{"$set": bson.M{"old": true}})

		fmt.Printf(".")
	}

	fmt.Println("Finished")
}

func (module Module) TestSocket() {

	carrier := module.Transmit

	carrierParams := map[string]interface{}{
		"fire":     "new-post",
		"category": "549da59c6461740097000000",
	}

	carrier.Emit("feed", "action", carrierParams)

	fmt.Println("feed action emmited")
}

func (module Module) ReplaceURL() {

	var usr user.UserPrivate
	var post model.Post

	database := module.Mongo.Database

	// Get all users
	iter := database.C("users").Find(nil).Iter()

	for iter.Next(&usr) {

		image := strings.Replace(usr.Image, "http://s3-us-west-1.amazonaws.com/spartan-board", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
		image = strings.Replace(usr.Image, "http://assets.spartangeek.com", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
		image = strings.Replace(usr.Image, "https://assets.spartangeek.com", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
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
		content = strings.Replace(post.Content, "http://assets.spartangeek.com", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
		content = strings.Replace(post.Content, "http://s3-us-west-1.amazonaws.com/spartan-board", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
		updates["content"] = content

		for index, comment := range post.Comments.Set {

			comment_index := strconv.Itoa(index)

			content := strings.Replace(comment.Content, "https://assets.spartangeek.com", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
			content = strings.Replace(comment.Content, "http://assets.spartangeek.com", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)
			content = strings.Replace(comment.Content, "http://s3-us-west-1.amazonaws.com/spartan-board", "https://s3-us-west-1.amazonaws.com/spartan-board", -1)

			updates["comments.set."+comment_index+".content"] = content
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

	var usr user.UserPrivate
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

func (module Module) MigrateDeletedComment() {

	var comment struct {
		Id       bson.ObjectId `bson:"_id"`
		Position int           `bson:"position"`
		Title    string        `bson:"title"`
		Comment  struct {
			UserId  bson.ObjectId `bson:"user_id"`
			Deleted time.Time     `bson:"deleted_at"`
		} `bson:"comment"`
	}

	database := module.Mongo.Database
	from, _ := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")

	// Get all users
	pipeline := database.C("posts_backup").Pipe([]bson.M{
		{
			"$unwind": bson.M{"path": "$comments.set", "includeArrayIndex": "position"},
		},
		{
			"$project": bson.M{"title": 1, "position": 1, "comment": "$comments.set"},
		},
		{
			"$match": bson.M{"comment.deleted_at": bson.M{"$gte": from}},
		},
	}).Iter()

	for pipeline.Next(&comment) {

		err := database.C("comments").Update(bson.M{"post_id": comment.Id, "user_id": comment.Comment.UserId, "position": comment.Position}, bson.M{"$set": bson.M{"deleted_at": comment.Comment.Deleted}})

		if err == nil {
			fmt.Printf(".")
		} else {
			fmt.Printf("-")
		}
	}
}

func (module Module) MigrateChosenComment() {

	var start string

	fmt.Println("Press enter to begin migration...")
	fmt.Scanln(&start)

	var comment struct {
		Id       bson.ObjectId `bson:"_id"`
		Position int           `bson:"position"`
		Title    string        `bson:"title"`
		Comment  struct {
			UserId bson.ObjectId `bson:"user_id"`
			Chosen bool          `bson:"chosen"`
		} `bson:"comment"`
	}

	database := module.Mongo.Database

	// Get all users
	pipeline := database.C("posts_backup").Pipe([]bson.M{
		{
			"$match": bson.M{"solved": true},
		},
		{
			"$unwind": bson.M{"path": "$comments.set", "includeArrayIndex": "position"},
		},
		{
			"$project": bson.M{"title": 1, "position": 1, "comment": "$comments.set"},
		},
		{
			"$match": bson.M{"comment.chosen": true},
		},
	}).Iter()

	for pipeline.Next(&comment) {

		err := database.C("comments").Update(bson.M{"post_id": comment.Id, "user_id": comment.Comment.UserId, "position": comment.Position}, bson.M{"$set": bson.M{"chosen": true}})

		if err == nil {
			fmt.Printf(".")
		} else {
			fmt.Printf("-")
		}
	}
}

func (module Module) Codes() {

	var usr user.UserPrivate
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

	var usr user.UserPrivate
	database := module.Mongo.Database

	// Get all users
	from := time.Now()
	from = from.Add(-time.Duration(time.Hour * 24 * 10))

	query := database.C("users").Find(bson.M{"validated": false, "gamificated_at": bson.M{"$gt": from}, "created_at": bson.M{"$lt": from}})
	count, _ := query.Count()

	fmt.Printf("Found %v at %v\n", count, from)

	iter := query.Iter()

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

			if !full_name_exists {

				full_name = name
			}

			image, image_exists := component["image"]

			if !image_exists {

				image = ""
			}

			id := component["_id"].(bson.ObjectId)

			if part_number, exists := component["part_number"]; exists {

				slug := component["slug"].(string)

				item := components.AlgoliaComponentModel{
					Id:       id.Hex(),
					Name:     name.(string),
					FullName: full_name.(string),
					Part:     part_number.(string),
					Slug:     slug,
					Image:    image.(string),
					Type:     component["type"].(string),
				}

				fmt.Printf("+")

				// Append to the current batch
				batch_count++
				batch_store = append(batch_store, item)

				batch_check(false)
			}
		}
	}

	// Upload last batch even if incomplete
	batch_check(true)

	fmt.Printf("*")
}
