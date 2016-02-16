package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/feed"
	"github.com/fernandez14/spartangeek-blacker/modules/transmit"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

	"strconv"
	"sort"
)

type PostAPI struct {
	Feed       *feed.FeedModule   `inject:""`
	Acl        *acl.Module        `inject:""`
	Components *components.Module `inject:""`
	Transmit   *transmit.Sender   `inject:""`
	Mongo      *mongo.Service     `inject:""`
}

func (this PostAPI) MarkCommentAsAnswer(c *gin.Context) {

	post_id := c.Param("id")
	comment_pos := c.Param("comment")

	if bson.IsObjectIdHex(post_id) == false {

		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	id := bson.ObjectIdHex(post_id)
	comment_index, err := strconv.Atoi(comment_pos)

	if err != nil {

		c.JSON(400, gin.H{"message": "Invalid request, comment index not valid.", "status": "error"})
		return
	}

	user_str_id := c.MustGet("user_id")
	user_id := bson.ObjectIdHex(user_str_id.(string))
	user := this.Acl.User(user_id)

	post, err := this.Feed.Post(id)

	if err != nil {

		c.JSON(404, gin.H{"status": "error", "message": "Post not found."})
		return
	}

	if user.CanSolvePost(post.Data()) == false {

		c.JSON(400, gin.H{"message": "Can't update post. Insufficient permissions", "status": "error"})
		return
	}

	if post.Data().Solved == true {

		c.JSON(400, gin.H{"status": "error", "message": "Already solved."})
		return
	}

	comment, err := post.Comment(comment_index)

	if err != nil {

		c.JSON(400, gin.H{"message": "Invalid request, comment index not valid.", "status": "error"})
		return
	}

	comment.MarkAsAnswer()

	go func(carrier *transmit.Sender, id bson.ObjectId) {

		carrierParams := map[string]interface{}{
			"fire": "best-answer",
			"id": id.Hex(),
		} 

		carrier.Emit("feed", "action", carrierParams)

	}(this.Transmit, post.Data().Id)

	c.JSON(200, gin.H{"status": "okay"})
}

func (this PostAPI) GetPostComments(c *gin.Context) {

	post_id := c.Param("id")
	offsetQuery := c.Query("offset")
	limitQuery := c.Query("limit")
	database := this.Mongo.Database

	var offset int = 0
	var limit int = 0
	var post model.CommentsPost

	if bson.IsObjectIdHex(post_id) == false {
		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	id := bson.ObjectIdHex(post_id)
	offsetC, err := strconv.Atoi(offsetQuery)

	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid request, offset not valid.", "status": "error"})
		return
	}

	limitC, err := strconv.Atoi(limitQuery)

	if err != nil || limitC <= 0 {
		c.JSON(400, gin.H{"message": "Invalid request, limit not valid.", "status": "error"})
		return
	}

	offset = offsetC
	limit = limitC

	err = database.C("posts").FindId(id).Select(bson.M{"_id": 1, "comments.set": bson.M{"$slice": []int{offset, limit}}}).One(&post)
	
	if err != nil {
		c.JSON(404, gin.H{"message": "Couldnt found post with that id.", "status": "error"})
		return
	}

	var users []model.User
	var users_ls []bson.ObjectId

	for _, c := range post.Comments.Set {
		users_ls = append(users_ls, c.UserId)
	}

	err = database.C("users").Find(bson.M{"_id": bson.M{"$in": users_ls}}).All(&users)

	if err != nil {
		panic(err)
	}

	usersMap := make(map[bson.ObjectId]interface{})

	var description string

	for _, user := range users {
		description = "Solo otro Spartan Geek más"

		if len(user.Description) > 0 {
			description = user.Description
		}

		usersMap[user.Id] = map[string]interface{}{
			"id":          user.Id.Hex(),
			"username":    user.UserName,
			"description": description,
			"image":       user.Image,
			"level":       user.Gaming.Level,
			"roles":       user.Roles,
		}
	}

	// Name of the set to get
	_, signed_in := c.Get("token")

	// Look for votes that has been already given
	var likes []model.Vote

	if signed_in {

		user_id := c.MustGet("user_id")
		user_bson_id := bson.ObjectIdHex(user_id.(string))

		// Get the likes given by the current user
		_ = database.C("votes").Find(bson.M{"type": "comment", "related_id": post.Id, "user_id": user_bson_id}).All(&likes)
	}	

	// This will calculate the position based on the sliced array
	count := 0

	if offset >= 0 {
		count = offset
	} else {
		count = this.Feed.TrueCommentCount(post.Id) + offset
	}

	for index := range post.Comments.Set {

		comment := &post.Comments.Set[index]

		// Save the position over the comment
		post.Comments.Set[index].Position = count + index

		// Check if user liked that comment already
		for _, vote := range likes {
			if vote.NestedType == strconv.Itoa(index) {
				post.Comments.Set[index].Liked = vote.Value
			}
		}

		if _, okay := usersMap[comment.UserId]; okay {
			post.Comments.Set[index].User = usersMap[comment.UserId]
		}
	}

	// Remove deleted comments from the set
	comments := post.Comments.Set[:0]

	for _, c := range post.Comments.Set {
		if c.Deleted.IsZero() == true {
			comments = append(comments, c)
		}
	}

	post.Comments.Set = comments

	// Sort by created at
	sort.Sort(model.ByCommentCreatedAt(post.Comments.Set))

	c.JSON(200, post)
}

func (this PostAPI) Relate(c *gin.Context) {

	post_id := c.Param("id")
	related_id := c.Param("related_id")

	if !bson.IsObjectIdHex(post_id) || !bson.IsObjectIdHex(related_id) {

		c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
		return
	}

	id := bson.ObjectIdHex(post_id)
	component_id := bson.ObjectIdHex(related_id)

	// Validate post and related component
	post, err := this.Feed.Post(id)

	if err != nil {

		c.JSON(404, gin.H{"status": "error", "message": "Post not found."})
		return
	}

	component, err := this.Components.Get(component_id)

	if err != nil {

		c.JSON(404, gin.H{"status": "error", "message": "Component not found."})
		return
	}

	post.Attach(component)

	c.JSON(200, gin.H{"status": "okay"})
}
