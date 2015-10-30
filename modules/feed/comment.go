package feed

import(
	"github.com/fernandez14/spartangeek-blacker/model"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

type Comment struct {
	post    *Post
	comment model.Comment
	index   int
}

func (self Comment) MarkAsAnswer() {

	// Get database instance
	database := self.post.DI().Mongo.Database
	id := self.post.Data().Id

	// Record position on set
	position := "comments.set." + strconv.Itoa(self.index) + ".chosen"

	// Update straight forward
	err := database.C("posts").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{position: true, "solved": true}})

	if err != nil {
		panic(err)
	}
}