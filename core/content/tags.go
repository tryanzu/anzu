package content

import (
	"regexp"

	"gopkg.in/mgo.v2/bson"
)

var tagRegex, _ = regexp.Compile(`(?i)\[([a-z0-9]+(:?))+\]`)
var tagParamsRegex, _ = regexp.Compile(`(?i)(([a-z0-9]+)(:?))+?`)

type tag struct {
	Original string
	Name     string
	Params   []string
}

type tags []tag

func (list tags) withTag(name string) tags {
	filtered := list[:0]
	for _, tag := range list {
		if tag.Name != name {
			continue
		}
		filtered = append(filtered, tag)
	}
	return filtered
}

func (list tags) getIdParams(index int) (id []bson.ObjectId) {
	for _, tag := range list {
		if len(tag.Params) < index+1 {
			continue
		}

		if cid := tag.Params[index]; bson.IsObjectIdHex(cid) {
			id = append(id, bson.ObjectIdHex(cid))
		}
	}
	return
}
