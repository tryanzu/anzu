package content

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"regexp"
	"strings"
)

var tag_regex, _ = regexp.Compile(`(?i)\[([a-z0-9]+(:?))+\]`)
var tag_params_regex, _ = regexp.Compile(`(?i)(([a-z0-9]+)(:?))+?`)

type Tag struct {
	Original string
	Name     string
	Params   []string
}

func (self Module) ParseMentionTags(o Parseable, tags []Tag) bool {

	if len(tags) > 0 {

		var ids []string
		c := o.GetContent()

		for _, tag := range tags {

			// Ensure tag for mentions and its params
			if tag.Name == "mention" && len(tag.Params) > 0 {
				if id := tag.Params[0]; bson.IsObjectIdHex(id) {
					ids = append(ids, id)
				}
			}
		}

		users := self.FetchUsersHelper(ids)

		for _, tag := range tags {

			// Ensure tag for mentions and its params
			if tag.Name == "mention" && len(tag.Params) > 0 {
				if id := tag.Params[0]; bson.IsObjectIdHex(id) {

					usr, exists := users[id]

					if exists {

						link := `<a class="user-mention" data-id="` + id + `" data-username="` + usr + `">@` + usr + `</a>`
						c = strings.Replace(c, tag.Original, link, -1)
					}
				}
			}
		}

		o.UpdateContent(c)
	}

	return true
}

func (self Module) FetchUsersHelper(ls []string) map[string]string {

	usrMap := map[string]string{}
	redis := self.Redis
	users, err := redis.HMGet("usernamesHash", ls...)

	if err == nil && len(users) > 0 {

		for index, id := range ls {

			// ls index should match users index
			usr := users[index]

			if len(usr) > 0 {
				usrMap[id] = string(usr)
			}
		}
	}

	missing := []bson.ObjectId{}

	for _, id := range ls {
		if _, exists := usrMap[id]; !exists {
			missing = append(missing, bson.ObjectIdHex(id))
		}
	}

	if len(missing) > 0 {

		var targets []struct {
			Id       bson.ObjectId `bson:"_id"`
			Username string        `bson:"username"`
		}

		database := self.Mongo.Database
		err := database.C("users").Find(bson.M{"_id": bson.M{"$in": missing}}).Select(bson.M{"username": 1}).All(&targets)

		if err == nil {
			for _, usr := range targets {
				usrMap[usr.Id.Hex()] = usr.Username

				go redis.HSet("usernamesHash", usr.Id.Hex(), usr.Username)
			}
		}
	}

	return usrMap
}

func (self Module) ParseTags(o Parseable) error {

	c := o.GetContent()
	tags := []Tag{}
	mtags := tag_regex.FindAllString(c, -1)

	for _, tag := range mtags {

		params := tag_params_regex.FindAllString(tag, -1)
		count := len(params) - 1

		for i, param := range params {
			if i != count {
				params[i] = param[:len(param)-1]
			}
		}

		// Check length of params JIC
		if len(params) > 0 {
			tags = append(tags, Tag{
				Original: tag,
				Name:     params[0],
				Params:   params[1:],
			})
		}
	}

	fmt.Printf("%v\n", tags)

	chain := []func(Parseable, []Tag) bool{
		self.ParseMentionTags,
	}

	for _, fn := range chain {
		next := fn(o, tags)

		if !next {
			break
		}
	}

	o.OnParseFinished()

	return nil
}
