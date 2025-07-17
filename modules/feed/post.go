package feed

import (
	"context"

	"github.com/tryanzu/core/board/votes"
	"github.com/tryanzu/core/core/content"
	"github.com/tryanzu/core/deps"
	"github.com/tryanzu/core/modules/helpers"
	"github.com/tryanzu/core/modules/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"html"
	"time"
)

// Post model refers to board posts
type Post struct {
	Id                primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Title             string               `bson:"title" json:"title"`
	Slug              string               `bson:"slug" json:"slug"`
	Type              string               `bson:"type" json:"type"`
	Content           string               `bson:"content" json:"content"`
	Categories        []string             `bson:"categories" json:"categories"`
	Category          primitive.ObjectID   `bson:"category" json:"category"`
	Comments          Comments             `bson:"comments" json:"comments"`
	Author            *user.UserSimple     `bson:"-" json:"author,omitempty"`
	UserId            primitive.ObjectID   `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Users             []primitive.ObjectID `bson:"users,omitempty" json:"users,omitempty"`
	Votes             votes.Votes          `bson:"votes" json:"votes"`
	RelatedComponents []primitive.ObjectID `bson:"related_components,omitempty" json:"related_components,omitempty"`
	Following         bool                 `bson:"following,omitempty" json:"following,omitempty"`
	Pinned            bool                 `bson:"pinned,omitempty" json:"pinned,omitempty"`
	Lock              bool                 `bson:"lock" json:"lock"`
	IsQuestion        bool                 `bson:"is_question" json:"is_question"`
	Solved            bool                 `bson:"solved,omitempty" json:"solved,omitempty"`
	Views             int                  `bson:"views,omitempty" json:"views"`
	Reached           int                  `bson:"reached,omitempty" json:"-"`
	Voted             []string             `bson:"-" json:"voted"`
	Created           time.Time            `bson:"created_at" json:"created_at"`
	Updated           time.Time            `bson:"updated_at" json:"updated_at"`
	Deleted           time.Time            `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`

	// Runtime generated pointers
	UsersHashtable map[string]interface{} `bson:"-" json:"usersHashtable"`
	di             *FeedModule
}

// Set Dependency Injection pointer
func (self *Post) SetDI(di *FeedModule) {
	self.di = di
}

func (self *Post) LoadUsersHashtables() {
	var users []user.UserSimple
	ids := self.Users
	ids = append(ids, self.UserId)

	if len(ids) == 0 {
		self.UsersHashtable = make(map[string]interface{})
		return
	}

	ctx := context.Background()
	collection := deps.Container.Mgo().Collection("users")
	cursor, err := collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &users)
	if err != nil {
		panic(err)
	}

	ht := make(map[string]interface{}, len(users))
	for _, u := range users {
		ht[u.UserName] = u
	}

	self.UsersHashtable = ht
}

// Comment loading by ID for post
func (self *Post) LoadCommentById(id primitive.ObjectID) error {

	var c *Comment

	// Use content module to run processors chain
	content := self.di.Content
	ctx := context.Background()
	database := deps.Container.Mgo()
	collection := database.Collection("comments")
	err := collection.FindOne(ctx, bson.M{"_id": id, "deleted_at": bson.M{"$exists": false}}).Decode(&c)

	if err != nil {
		self.Comments.Set = make([]*Comment, 0)
		return err
	}

	self.Comments.Set = []*Comment{c}

	for _, comment := range self.Comments.Set {
		_ = content.ParseTags(comment)
	}

	return nil
}

// Push Comment on the post
func (self *Post) PushComment(c string, user_id primitive.ObjectID) *Comment {

	c = html.EscapeString(c)
	if len(c) > 3000 {
		c = c[:3000] + "..."
	}

	pos := self.GetCommentCount()

	comment := &Comment{
		Id:       primitive.NewObjectID(),
		PostId:   self.Id,
		UserId:   user_id,
		Content:  c,
		Position: pos,
		Created:  time.Now(),
	}

	comment.SetDI(self)

	// Use content module to run processors chain
	content := self.di.Content
	_ = content.Parse(comment)

	// Publish comment
	ctx := context.Background()
	database := deps.Container.Mgo()
	_, err := database.Collection("comments").InsertOne(ctx, comment)
	if err != nil {
		panic(err)
	}

	_, err = database.Collection("posts").UpdateOne(ctx, bson.M{"_id": self.Id}, bson.M{"$set": bson.M{"updated_at": time.Now()}, "$inc": bson.M{"comments.count": 1}})
	if err != nil {
		panic(err)
	}

	// Ensure user is participating
	self.PushUser(user_id)

	// Finally parse tags in content for runtime usage
	_ = content.ParseTags(comment)

	return comment
}

// Push new user to the participants list
func (p *Post) PushUser(user_id primitive.ObjectID) bool {

	pushed := false

	for _, u := range p.Users {

		if u == user_id {
			pushed = true
		}
	}

	if !pushed {
		ctx := context.Background()
		_, err := deps.Container.Mgo().Collection("posts").UpdateOne(ctx, bson.M{"_id": p.Id}, bson.M{"$push": bson.M{"users": user_id}})

		if err != nil {
			panic(err)
		}
	}

	return pushed
}

// LoadUsers for a post.
func (p *Post) LoadUsers() {

	var list []primitive.ObjectID
	var users []user.UserSimple

	// Check if author need to be loaded
	if p.Author == nil {
		list = append(list, p.UserId)
	}

	// Load comment set authors at runtime
	if len(p.Comments.Set) > 0 {
		for _, c := range p.Comments.Set {

			// Do not repeat ids at the list
			if exists, _ := helpers.InArray(c.UserId, list); !exists {
				list = append(list, c.UserId)
			}
		}

		// Best answer author
		if p.Comments.Answer != nil {
			if exists, _ := helpers.InArray(p.Comments.Answer.UserId, list); !exists {
				list = append(list, p.Comments.Answer.UserId)
			}
		}
	}

	if len(list) > 0 {
		ctx := context.Background()
		collection := deps.Container.Mgo().Collection("users")
		cursor, err := collection.Find(ctx, bson.M{"_id": bson.M{"$in": list}}, &options.FindOptions{
			Projection: user.UserSimpleFields,
		})
		if err != nil {
			panic(err)
		}
		defer cursor.Close(ctx)
		err = cursor.All(ctx, &users)
		if err != nil {
			panic(err)
		}

		usersMap := make(map[primitive.ObjectID]interface{})

		for i, usr := range users {
			usersMap[usr.Id] = usr

			if usr.Id == p.UserId {
				p.Author = &users[i]
			}

			if p.Comments.Answer != nil && p.Comments.Answer.UserId == usr.Id {
				p.Comments.Answer.User = usersMap[usr.Id]
			}
		}

		for index, c := range p.Comments.Set {
			if usr, exists := usersMap[c.UserId]; exists {
				p.Comments.Set[index].User = usr
			}
		}
	}
}

// LoadVotes for a post/user.
func (p *Post) LoadVotes(user_id primitive.ObjectID) {
	var list []votes.Vote
	ctx := context.Background()
	collection := deps.Container.Mgo().Collection("votes")
	cursor, err := collection.Find(ctx, bson.M{
		"type":       "post",
		"related_id": p.Id,
		"user_id":    user_id,
		"deleted_at": bson.M{"$exists": false},
	})
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &list)
	if err != nil {
		panic(err)
	}
	p.Voted = make([]string, 0)
	for _, v := range list {
		p.Voted = append(p.Voted, v.Value)
	}
}

// Get post data structure
func (self *Post) Data() *Post {
	return self
}

func (self *Post) IsLocked() bool {
	return self.Lock
}

func (self *Post) DI() *FeedModule {
	return self.di
}

// Alias of TrueCommentCount
func (self *Post) GetCommentCount() int {
	return self.di.TrueCommentCount(self.Id)
}

func (p *Post) GetContent() string {
	return p.Content
}

func (p *Post) UpdateContent(content string) content.Parseable {
	p.Content = content
	return p
}

func (p *Post) GetParseableMeta() map[string]interface{} {
	meta := make(map[string]interface{})
	meta["id"] = p.Id
	meta["related"] = "comment"
	meta["user_id"] = p.UserId
	return meta
}
