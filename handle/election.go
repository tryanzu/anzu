package handle

import (
	"bytes"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type ElectionAPI struct {
	DataService *mongo.Service `inject:""`
}

func (di *ElectionAPI) ElectionAddOption(c *gin.Context) {

	// Get the database interface from the DI
	database := di.DataService.Database

	id := c.Params.ByName("id")

	if bson.IsObjectIdHex(id) == false {

		// Invalid request
		c.JSON(400, gin.H{"error": "Invalid request...", "status": 507})

		return
	}

	// Get the query parameters
	qs := c.Request.URL.Query()

	// Name of the set to get
	token := qs.Get("token")

	if token == "" {

		c.JSON(401, gin.H{"error": "Not authorized...", "status": 502})
		return
	}

	// Get user by token
	user_token := model.UserToken{}

	// Try to fetch the user using token header though
	err := database.C("tokens").Find(bson.M{"token": token}).One(&user_token)

	if err != nil {

		c.JSON(401, gin.H{"error": "Not authorized...", "status": 503})

		return
	}

	var option model.ElectionForm

	if c.BindWith(&option, binding.JSON) == nil {

		// Check if component is valid
		component := option.Component
		content := option.Content
		valid := false

		for _, possible := range avaliable_components {

			if component == possible {

				valid = true
			}
		}

		if valid == true {

			// Posts collection
			collection := database.C("posts")

			var post model.Post

			err := collection.FindId(id).One(&post)

			if err != nil {

				// No guest can vote
				c.JSON(404, gin.H{"error": "Couldnt found post with that id.", "status": 501})

				return
			}

			votes := model.Votes{
				Up:   0,
				Down: 0,
			}

			election := model.ElectionOption{
				UserId:  user_token.UserId,
				Content: content,
				Votes:   votes,
				Created: time.Now(),
			}

			var push bytes.Buffer

			// Make the push string
			push.WriteString("components.")
			push.WriteString(component)
			push.WriteString(".options")

			over := push.String()

			change := bson.M{"$push": bson.M{over: election}, "$set": bson.M{"updated_at": time.Now()}}
			err = collection.Update(bson.M{"_id": post.Id}, change)

			if err != nil {
				panic(err)
			}

			// Check if we need to add participant
			users := post.Users
			need_add := true

			for _, already_within := range users {

				if already_within == user_token.UserId {

					need_add = false
				}
			}

			if need_add == true {

				// Add the user to the user list
				change := bson.M{"$push": bson.M{"users": user_token.UserId}}
				err = collection.Update(bson.M{"_id": post.Id}, change)

				if err != nil {
					panic(err)
				}
			}

			c.JSON(200, gin.H{"message": "okay", "status": 506})
			return
		}
	}

	c.JSON(401, gin.H{"error": "Not authorized", "status": 504})
}
