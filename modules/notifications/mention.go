package notifications

import (
	"fmt"
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"gopkg.in/mgo.v2/bson"
	"regexp"
	"strings"
	"time"
)

func (di *NotificationsModule) ParseContentMentions(obj MentionParseObject) {

	// Recover from any panic even inside this isolated process
	defer di.Errors.Recover()

	var mention_users, mentions_done []string
	var content string
	var author model.User
	var post model.Post

	titles := map[string]string{
		"comment": "**%s** te mencionó en un comentario",
		"post":    "**%s** te mencionó en su publicación",
	}

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
			var target_username, title, message, link string

			target_username = username

			err := database.C("users").Find(bson.M{"username": target_username}).One(&target_user)

			// Don't send the notification if the user has not been found or if the target is the same as the author
			if err != nil || target_user.Id == obj.Author || target_user.Id == post.UserId  {
				continue
			}

			// Replace the mention in the content so it can be a link to the profile
			if comment_index != "" {

				link = `<a class="user-mention" data-id="` + target_user.Id.Hex() + `" data-username="` + target_username + `" data-comment="` + comment_index + `">@` + target_username + `</a>`
			} else {

				link = `<a class="user-mention" data-id="` + target_user.Id.Hex() + `" data-username="` + target_username + `">@` + target_username + `</a>`
			}

			content = strings.Replace(content, user, link, -1)

			title = fmt.Sprintf(titles[obj.Type], author.UserName)
			message = obj.Title

			// Compose notification
			notification := &model.UserFirebaseNotification{
				UserId:       obj.Author,
				RelatedId:    post.Id,
				RelatedExtra: post.Slug,
				Position:     post.Comments.Count,
				Title:        title,
				Text:         message,
				Related:      "mention",
				Seen:         false,
				Image:        "",
				Created:      time.Now(),
				Updated:      time.Now(),
			}
			
			// Send using the broadcaster
			broadcaster.Send(notification)

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
