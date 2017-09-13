package posts

import (
	"github.com/fernandez14/spartangeek-blacker/core/events"
	"github.com/fernandez14/spartangeek-blacker/core/user"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/mgo.v2/bson"

	"html"
	"reflect"
	"regexp"
	"time"
)

func (this API) Create(c *gin.Context) {
	var form model.PostForm
	database := this.Mongo.Database

	// Check for user token
	user_id, _ := c.Get("user_id")
	bson_id := bson.ObjectIdHex(user_id.(string))

	// Get the form otherwise tell it has been an error
	if c.BindWith(&form, binding.JSON) == nil {

		post_category := form.Category

		if bson.IsObjectIdHex(post_category) == false {
			c.JSON(400, gin.H{"status": "error", "message": "Invalid category id"})
			return
		}

		var category model.Category

		err := database.C("categories").Find(bson.M{"parent": bson.M{"$exists": true}, "_id": bson.ObjectIdHex(post_category)}).One(&category)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": "Invalid category"})
			return
		}

		usr := this.Acl.User(bson_id)

		if usr.CanWrite(category) == false || usr.HasValidated() == false {
			c.JSON(403, gin.H{"status": "error", "message": "Not enough permissions."})
			return
		}

		if form.Pinned == true && usr.Can("pin-board-posts") == false {
			c.JSON(400, gin.H{"status": "error", "message": "Not enough permissions to pin."})
			return
		}

		if form.Lock == true && usr.Can("block-own-post-comments") == false {
			c.JSON(400, gin.H{"status": "error", "message": "Not enough permissions to lock."})
			return
		}

		comments := model.Comments{
			Count: 0,
			Set:   make([]model.Comment, 0),
		}

		votes := model.Votes{
			Up:     0,
			Down:   0,
			Rating: 0,
		}

		content := html.EscapeString(form.Content)
		urls, _ := regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
		post_id := bson.NewObjectId()

		var assets []string

		assets = urls.FindAllString(content, -1)

		// Empty participants list - only author included
		users := []bson.ObjectId{bson_id}

		switch form.Kind {
		case "recommendations":
			components := form.Components

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

				// Some important information is missing for this kind of post
				if !bo || !bto || !bco || !bfo || !so {
					c.JSON(400, gin.H{"status": "error", "message": "Couldnt create post, missing information...", "code": 4001})
					return
				}

				post_name := "PC '" + form.Name
				if budget.(string) != "0" {
					post_name += "' con presupuesto de $" + budget.(string) + " " + budget_currency.(string)
				} else {
					post_name += "'"
				}

				slug := helpers.StrSlug(post_name)
				slug_exists, _ := database.C("posts").Find(bson.M{"slug": slug}).Count()

				if slug_exists > 0 {

					slug = helpers.StrSlugRandom(post_name)
				}

				publish := &model.Post{
					Id:         post_id,
					Title:      post_name,
					Content:    content,
					Type:       "recommendations",
					Slug:       slug,
					Comments:   comments,
					UserId:     bson_id,
					Users:      users,
					Categories: []string{"recommendations"},
					Category:   bson.ObjectIdHex(post_category),
					Votes:      votes,
					IsQuestion: form.IsQuestion,
					Pinned:     form.Pinned,
					Lock:       form.Lock,
					Created:    time.Now(),
					Updated:    time.Now(),
				}

				publish_components := model.Components{
					Budget:            budget.(string),
					BudgetType:        budget_type.(string),
					BudgetCurrency:    budget_currency.(string),
					BudgetFlexibility: budget_flexibility.(string),
					Software:          software.(string),
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

						if name == component || name == component+",omitempty" {

							c := model.Component{
								Elections: component_elements["poll"].(bool),
								Status:    status,
								Votes:     votes,
								Content:   component_elements["value"].(string),
							}

							// Set the component with the component we've build above
							bindable.Field(i).Set(reflect.ValueOf(c))
						}
					}
				}

				// Now bind the components to the post
				publish.Components = publish_components

				u, err := user.FindId(deps.Container, bson_id)
				if err != nil {
					c.AbortWithError(500, err)
					return
				}

				if user.CanBeTrusted(u) == false {
					publish.Deleted = time.Now()
				}

				err = database.C("posts").Insert(publish)

				if err != nil {
					panic(err)
				}

				for _, asset := range assets {

					// Download the asset on other routine in order to non block the API request
					go this.savePostImages(asset, publish.Id)
				}

				// Notify events pool.
				events.In <- events.PostNew(publish.Id)

				// Finished creating the post
				c.JSON(200, gin.H{"status": "okay", "code": 200, "post": gin.H{"id": post_id, "slug": slug}})
				return
			}

		case "category-post":

			title := form.Name

			if len([]rune(title)) > 72 {
				title = helpers.Truncate(title, 72) + "..."
			}

			slug := helpers.StrSlug(title)
			slug_exists, _ := database.C("posts").Find(bson.M{"slug": slug}).Count()

			if slug_exists > 0 {
				slug = helpers.StrSlugRandom(title)
			}

			publish := &model.Post{
				Id:         post_id,
				Title:      title,
				Content:    content,
				Type:       "category-post",
				Slug:       slug,
				Comments:   comments,
				UserId:     bson_id,
				Users:      users,
				Category:   bson.ObjectIdHex(post_category),
				Votes:      votes,
				IsQuestion: form.IsQuestion,
				Pinned:     form.Pinned,
				Lock:       form.Lock,
				Created:    time.Now(),
				Updated:    time.Now(),
			}

			u, err := user.FindId(deps.Container, bson_id)
			if err != nil {
				c.AbortWithError(500, err)
				return
			}

			if user.CanBeTrusted(u) == false {
				publish.Deleted = time.Now()
			}

			err = database.C("posts").Insert(publish)
			if err != nil {
				panic(err)
			}

			// Notify events pool immediately after performing save.
			events.In <- events.PostNew(publish.Id)

			for _, asset := range assets {

				// Non blocking image download
				go this.savePostImages(asset, publish.Id)
			}

			// Finished creating the post
			c.JSON(200, gin.H{"status": "okay", "code": 200, "post": gin.H{"id": post_id, "slug": slug}})
			return
		}
	}

	c.JSON(400, gin.H{"status": "error", "message": "Couldnt create post, missing information...", "code": 205})
}
