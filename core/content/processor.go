package content

import (
	"log"
	"strings"
	"time"

	"github.com/fernandez14/spartangeek-blacker/core/common"
	"github.com/fernandez14/spartangeek-blacker/core/user"
	"gopkg.in/mgo.v2/bson"
)

// Content processor definition.
type Processor func(Deps, Parseable, tags) (Parseable, error)

// Postprocess a parseable type.
func Postprocess(deps Deps, c Parseable) (processed Parseable, err error) {
	starts := time.Now()
	list := parseTags(c)
	pipeline := []Processor{
		replaceMentionTags,
	}

	// Run pipeline over parseable.
	processed = c
	for _, fn := range pipeline {
		processed, err = fn(deps, processed, list)

		if err != nil {
			return
		}
	}

	elapsed := time.Since(starts)
	log.Printf("Parsable postprocess took: %v\n", elapsed)
	return
}

// Replace mention related tags with links to mentioned user.
func replaceMentionTags(deps Deps, c Parseable, list tags) (processed Parseable, err error) {
	processed = c
	if len(list) == 0 {
		return
	}

	mentions := list.withTag("mention")
	usersId := mentions.getIdParams(0)
	if len(usersId) == 0 {
		return
	}

	var users common.UsersStringMap
	users, err = user.FindNames(deps, usersId...)
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

			link := `<a class="user-mention" data-id="` + id + `" data-username="` + name + `">@` + name + `</a>`
			content = strings.Replace(content, tag.Original, link, -1)
		}
	}

	processed = processed.UpdateContent(content)
	return
}

func parseTags(c Parseable) (list []tag) {

	// Use regex to find all tags inside the parseable content.
	found := tagRegex.FindAllString(c.GetContent(), -1)
	for _, match := range found {

		// Having parsed all tags now destructure the tag params.
		params := tagParamsRegex.FindAllString(match, -1)
		count := len(params) - 1

		for n, param := range params {
			if n != count {
				params[n] = param[:len(param)-1]
			}
		}

		if len(params) > 0 {
			list = append(list, tag{match, params[0], params[1:]})
		}
	}
	return
}
