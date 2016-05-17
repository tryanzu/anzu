package notifications

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"gopkg.in/mgo.v2/bson"

	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (self *NotificationsModule) Mention(parseableMeta map[string]interface{}, user_id, target_user bson.ObjectId) {

	defer self.Errors.Recover()

	usr, err := self.User.Get(user_id)

	if err != nil {
		panic(fmt.Sprintf("Could not get user while notifying mention (user_id: %v, target_user: %v). It said: %v", user_id, target_user, err))
	}

	position, exists := parseableMeta["position"].(int)

	if !exists {
		panic(fmt.Sprintf("Position does not exists in parseable meta (%v)", parseableMeta))
	}

	post, exists := parseableMeta["post"].(map[string]interface{})

	if !exists {
		panic(fmt.Sprintf("Post does not exists in parseable meta (%v)", parseableMeta))
	}

	post_id, exists := post["id"].(bson.ObjectId)

	if !exists {
		panic(fmt.Sprintf("post_id does not exists in parseable meta (%v)", parseableMeta))
	}

	post_slug, exists := post["slug"].(string)

	if !exists {
		panic(fmt.Sprintf("post_slug does not exists in parseable meta (%v)", parseableMeta))
	}

	post_title, exists := post["title"].(string)

	if !exists {
		panic(fmt.Sprintf("post_title does not exists in parseable meta (%v)", parseableMeta))
	}

	notification := model.UserFirebaseNotification{
		UserId:       target_user,
		RelatedId:    post_id,
		RelatedExtra: post_slug,
		Position:     strconv.Itoa(position),
		Username:     usr.Name(),
		Text:         post_title,
		Related:      "mention",
		Seen:         false,
		Image:        usr.Data().Image,
		Created:      time.Now(),
		Updated:      time.Now(),
	}

	broadcaster := self.Broadcaster
	broadcaster.Send(notification)

}

func (di *NotificationsModule) ParseContentMentions(obj MentionParseObject) {

	// Recover from any panic even inside this isolated process
	defer di.Errors.Recover()

	var mention_users, mentions_done []string
	var content string
	var author model.User
	var post model.Post

	// This is the mention regex to determine the possible users to mention
	mention_regex, _ := regexp.Compile(`(?i)\B\@([\w\-]+)(#[0-9]+)*`)
	mention_comment_regex, _ := regexp.Compile(`(?i)\B\@([\w\-]+)#[0-9]+`)

	// Services we will need along the runtime
	database := di.Mongo.Database
	broadcaster := di.Broadcaster

	post = obj.Post
	content = obj.Content
	mention_users = mention_regex.FindAllString(content, -1)

	// Get the author of the notification
	err := database.C("users").Find(bson.M{"_id": obj.Author}).One(&author)

	if err != nil {
		return
	}

	for _, user := range mention_users {

		var username string
		var comment_index string

		if mention_comment_regex.MatchString(user) {

			// Split the parts of the mention
			mentions_parts := mention_regex.FindStringSubmatch(user)
			username = mentions_parts[1]
			comment_index = mentions_parts[2][1:]

		} else {

			username = user[1:]
		}

		if done, _ := helpers.InArray(username, mentions_done); done == false {

			var target_user model.User
			var target_username, message, link string

			target_username = username

			err := database.C("users").Find(bson.M{"username": target_username}).One(&target_user)

			// Don't send the notification if the user has not been found or if the target is the same as the author
			if err != nil || target_user.Id == obj.Author {
				continue
			}

			// Replace the mention in the content so it can be a link to the profile
			if comment_index != "" {

				link = `<a class="user-mention" data-id="` + target_user.Id.Hex() + `" data-username="` + target_username + `" data-comment="` + comment_index + `">@` + target_username + `</a>`
			} else {

				link = `<a class="user-mention" data-id="` + target_user.Id.Hex() + `" data-username="` + target_username + `">@` + target_username + `</a>`
			}

			content = strings.Replace(content, user, link, -1)
			message = obj.Title

			if target_user.Id != post.UserId {

				position, err := strconv.Atoi(obj.RelatedNested)

				if err != nil {
					panic(err)
				}

				// Check if mention has been sent already before for this same entity
				sent, err := database.C("mentions").Find(bson.M{"post_id": post.Id, "user_id": target_user.Id, "nested": position}).Count()

				if err != nil || sent > 0 {
					continue
				}

				// Compose notification
				notification := model.UserFirebaseNotification{
					UserId:       target_user.Id,
					RelatedId:    post.Id,
					RelatedExtra: post.Slug,
					Position:     strconv.Itoa(position),
					Username:     author.UserName,
					Text:         message,
					Related:      "mention",
					Seen:         false,
					Image:        author.Image,
					Created:      time.Now(),
					Updated:      time.Now(),
				}

				mention := &model.MentionModel{
					PostId: post.Id,
					UserId: target_user.Id,
					Nested: position,
				}

				// Insert the mention for further uses
				database.C("mentions").Insert(mention)

				// Send using the broadcaster
				broadcaster.Send(notification)
			}

			// Dont send repeated notifications to the same user even if mentioned twice
			mentions_done = append(mentions_done, target_username)
		}
	}

	if obj.Type == "comment" {

		path := "comments.set." + obj.RelatedNested + ".content"

		// Compute the database change directives
		change := bson.M{"$set": bson.M{"updated_at": time.Now(), path: content}}

		err = database.C("posts").Update(bson.M{"_id": post.Id}, change)

		if err != nil {
			return
		}
	}
}
