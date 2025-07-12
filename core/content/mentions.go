package content

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/tryanzu/core/core/common"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	mentions, _       = regexp.Compile(`(?i)\s?\B\@([\w\-]+)(#[0-9]+)*`)
	commentMention, _ = regexp.Compile(`(?i)\B\@([\w\-]+)#[0-9]+`)
)

func preReplaceMentionTags(d deps, c Parseable) (processed Parseable, err error) {
	processed = c
	content := processed.GetContent()
	list := mentions.FindAllString(content, -1)
	if len(list) == 0 {
		return
	}

	var users []string
	possible := map[string]Mention{}

	for _, usr := range list {
		var (
			username string
			cidx     string
		)

		// Split the parts of the mention
		if commentMention.MatchString(usr) {
			parts := mentions.FindStringSubmatch(usr)
			username = parts[1]
			cidx = parts[2][1:]
		} else {
			usr = strings.TrimSpace(usr)
			username = usr[1:]
		}

		if _, done := possible[username]; done {
			continue
		}

		users = append(users, username)
		possible[username] = Mention{
			Username: username,
			Comment:  cidx,
			Original: usr,
		}
	}

	var targets []struct {
		ID       primitive.ObjectID `bson:"_id"`
		Username string             `bson:"username"`
	}

	if len(users) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := d.Mgo().Collection("users").Find(ctx, bson.M{"username": bson.M{"$in": users}})
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &targets)
	if err != nil || len(targets) == 0 {
		return
	}

	meta := processed.GetParseableMeta()
	relatedID := meta["id"].(primitive.ObjectID)
	related := meta["related"].(string)
	userID := meta["user_id"].(primitive.ObjectID)
	usersID := []primitive.ObjectID{userID}

	var refs []Mention
	for _, usr := range targets {
		mention, exists := possible[usr.Username]
		if exists == false {
			continue
		}
		mention.UserID = usr.ID
		refs = append(refs, mention)
		content = mention.Replace(content)

		// Track mention
		events.In <- events.TrackMention(usr.ID, relatedID, related, usersID)
	}

	processed = processed.UpdateContent(content)
	return
}

// Replace mention related tags with links to mentioned user.
func postReplaceMentionTags(d deps, c Parseable, list tags) (processed Parseable, err error) {
	processed = c
	if len(list) == 0 {
		return
	}

	mentions := list.withTag("mention")
	usersID := mentions.getIdParams(0)
	if len(usersID) == 0 {
		return
	}

	var users common.UsersStringMap
	users, err = user.FindNames(d, usersID...)
	if err != nil {
		return
	}

	content := processed.GetContent()
	for _, tag := range mentions {
		if id := tag.Params[0]; primitive.IsValidObjectID(id) {
			oidHex, _ := primitive.ObjectIDFromHex(id)
			name, exists := users[oidHex]
			if exists == false {
				continue
			}

			link := `[@` + name + `](/u/` + name + `/` + id + `)`
			content = strings.Replace(content, tag.Original, link, -1)
		}
	}

	processed = processed.UpdateContent(content)
	return
}

// Mention ref.
type Mention struct {
	UserID   primitive.ObjectID
	Username string
	Comment  string
	Original string
}

func (m Mention) Replace(content string) string {
	tag := "[mention:" + m.UserID.Hex()
	if m.Comment != "" {
		tag = tag + ":" + m.Comment
	}
	tag = tag + "]"

	return strings.Replace(content, m.Original, tag, -1)
}
