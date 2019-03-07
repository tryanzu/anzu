package content

import (
	"regexp"
	"strings"

	"github.com/tryanzu/core/core/common"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/user"
	"gopkg.in/mgo.v2/bson"
)

var (
	mentions, _       = regexp.Compile(`(?i)\B\@([\w\-]+)(#[0-9]+)*`)
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
		ID       bson.ObjectId `bson:"_id"`
		Username string        `bson:"username"`
	}

	err = d.Mgo().C("users").Find(bson.M{"username": bson.M{"$in": users}}).Select(bson.M{"username": 1}).All(&targets)
	if err != nil || len(targets) == 0 {
		return
	}

	meta := processed.GetParseableMeta()
	relatedID := meta["id"].(bson.ObjectId)
	userID := meta["user_id"].(bson.ObjectId)
	usersID := []bson.ObjectId{userID}

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
		events.In <- events.TrackMention(usr.ID, relatedID, usersID)
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
		if id := tag.Params[0]; bson.IsObjectIdHex(id) {
			name, exists := users[bson.ObjectIdHex(id)]
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
	UserID   bson.ObjectId
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
