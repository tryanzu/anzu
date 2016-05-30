package content

import (
	"gopkg.in/mgo.v2/bson"

	"regexp"
	"strings"
)

var mention_regex, _ = regexp.Compile(`(?i)\B\@([\w\-]+)(#[0-9]+)*`)
var mention_comment_regex, _ = regexp.Compile(`(?i)\B\@([\w\-]+)#[0-9]+`)

type Mention struct {
	UserId   bson.ObjectId
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
		database := self.Mongo.Database

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
					bson.ObjectId(""),
					username,
					comment_index,
					usr,
				}
			}
		}

		var targets []struct {
			Id       bson.ObjectId `bson:"_id"`
			Username string        `bson:"username"`
		}

		var mentions []Mention

		err := database.C("users").Find(bson.M{"username": bson.M{"$in": users}}).Select(bson.M{"username": 1}).All(&targets)

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

	defer self.Errors.Recover()

	database := self.Mongo.Database
	entity := o.GetParseableMeta()

	if related, exists := entity["type"].(string); exists {

		if related_id, exists := entity["id"].(bson.ObjectId); exists {

			var owner_id bson.ObjectId
			var post_owner bson.ObjectId

			if oid, exists := entity["owner_id"]; exists {
				owner_id = oid.(bson.ObjectId)
			}

			if p, exists := entity["post"].(map[string]interface{}); exists {
				if uid, exists := p["user_id"]; exists {
					post_owner = uid.(bson.ObjectId)
				}
			}

			for _, to := range ls {

				// If owner_id is valid don't notify the owner
				if owner_id.Valid() && owner_id == to.UserId {
					continue
				}

				// Ignore mentions to post_owner (if any)
				if post_owner.Valid() && post_owner == to.UserId {
					continue
				}

				sent, err := database.C("mentions").Find(bson.M{"related": related, "related_id": related_id, "user_id": to.UserId}).Count()

				// Check if mention has been sent already before for this same entity
				if err != nil || sent > 0 {
					continue
				}

				self.Notifications.Mention(entity, owner_id, to.UserId)
			}

		} else {
			panic("NotifyMentionsAsync could not get parseable metadata id")
		}

	} else {
		panic("NotifyMentionsAsync could not get parseable metadata type")
	}
}
