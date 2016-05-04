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
	list = mention_regex.FindAllString(c, -1)

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
					bson.ObjectId{},
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

					mention.Id = usr.Id

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

		// Async mentions notifying
		go self.NotifyMentionsAsync(mentions)
	}

	o.OnParseFilterFinished("mentions")
}

func (self Module) NotifyMentionsAsync(ls []Mention) {

	defer self.Errors.Recover()

}
