package content

import (
	"context"
	"regexp"
	"strings"

	notify "github.com/tryanzu/core/board/notifications"
	"github.com/tryanzu/core/deps"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mention_regex, _ = regexp.Compile(`(?i)\B\@([\w\-]+)(#[0-9]+)*`)
var mention_comment_regex, _ = regexp.Compile(`(?i)\B\@([\w\-]+)#[0-9]+`)

type Mention struct {
	UserId   primitive.ObjectID
	Username string
	Comment  string
	Original string
}

func (self Module) ParseContentMentions(o Parseable) bool {

	c := o.GetContent()
	list := mention_regex.FindAllString(c, -1)

	if len(list) > 0 {

		var users []string
		possible := map[string]Mention{}
		database := deps.Container.Mgo()

		for _, usr := range list {
			var username string
			var comment_index string

			if mention_comment_regex.MatchString(usr) {
				// Split the parts of the mention
				mentions_parts := mention_regex.FindStringSubmatch(usr)
				username = mentions_parts[1]
				comment_index = mentions_parts[2][1:]
			} else {
				username = usr[1:]
			}

			if _, done := possible[username]; !done {
				users = append(users, username)
				possible[username] = Mention{
					primitive.NilObjectID,
					username,
					comment_index,
					usr,
				}
			}
		}

		var targets []struct {
			Id       primitive.ObjectID `bson:"_id"`
			Username string            `bson:"username"`
		}

		var mentions []Mention

		ctx := context.Background()
		collection := database.Collection("users")
		filter := bson.M{"username": bson.M{"$in": users}}
		opts := options.Find().SetProjection(bson.M{"username": 1})
		cursor, err := collection.Find(ctx, filter, opts)
		if err != nil {
			return false
		}
		defer cursor.Close(ctx)
		err = cursor.All(ctx, &targets)
		if err != nil {
			return false
		}

		if err == nil && len(targets) > 0 {
			for _, usr := range targets {
				if mention, exists := possible[usr.Username]; exists {
					mention.UserId = usr.Id
					tag := "[mention:" + usr.Id.Hex() + "]"
					if mention.Comment != "" {
						tag = "[mention:" + usr.Id.Hex() + ":" + mention.Comment + "]"
					}
					mentions = append(mentions, mention)
					c = strings.Replace(c, mention.Original, tag, -1)
				}
			}
		}

		o.UpdateContent(c)

		// Asynchronously mentions notifying
		go self.NotifyMentionsAsync(o, mentions)
	}

	o.OnParseFilterFinished("mentions")

	return true
}

func (self Module) NotifyMentionsAsync(o Parseable, ls []Mention) {
	ctx := context.Background()
	defer self.Errors.Recover()

	database := deps.Container.Mgo()
	entity := o.GetParseableMeta()

	if related, exists := entity["type"].(string); exists {
		if related_id, exists := entity["id"].(primitive.ObjectID); exists {
			var owner_id primitive.ObjectID
			var post_owner primitive.ObjectID

			if oid, exists := entity["owner_id"]; exists {
				owner_id = oid.(primitive.ObjectID)
			}

			if p, exists := entity["post"].(map[string]interface{}); exists {
				if uid, exists := p["user_id"]; exists {
					post_owner = uid.(primitive.ObjectID)
				}
			}

			for _, to := range ls {
				// If owner_id is valid don't notify the owner
				if !owner_id.IsZero() && owner_id == to.UserId {
					continue
				}

				// Ignore mentions to post_owner (if any)
				if !post_owner.IsZero() && post_owner == to.UserId {
					continue
				}

				mentionsCollection := database.Collection("mentions")
				filter := bson.M{"related": related, "related_id": related_id, "user_id": to.UserId}
				sent, err := mentionsCollection.CountDocuments(ctx, filter)

				// Check if mention has been sent already before for this same entity
				if err != nil || sent > 0 {
					continue
				}

				notify.Database <- notify.Notification{
					UserId:    to.UserId,
					Type:      "mention",
					RelatedId: related_id,
					Users:     []primitive.ObjectID{owner_id},
				}

				// TODO: persist new mentions here instead...
				self.Notifications.Mention(entity, owner_id, to.UserId)
			}
		} else {
			panic("NotifyMentionsAsync could not get parseable metadata id")
		}
	} else {
		panic("NotifyMentionsAsync could not get parseable metadata type")
	}
}
