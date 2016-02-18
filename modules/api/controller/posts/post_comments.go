package posts

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

	"strconv"
	"sort"
)

func (this API) GetPostComments(c *gin.Context) {

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

	// This will calculate the position based on the sliced array
	count := 0
	true_count := this.Feed.TrueCommentCount(id)

	if offset >= 0 {
		count = offset
	} else {

		offsetN := -offset 

		if offsetN  > true_count {
			offset = -true_count
		}

		count = true_count + offset
	}

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
