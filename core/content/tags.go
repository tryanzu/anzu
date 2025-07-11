package content

import (
	"regexp"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
	filtered := tags{}
	for _, tag := range list {
		if tag.Name != name {
			continue
		}
		filtered = append(filtered, tag)
	}
	return filtered
}

func (list tags) getIdParams(index int) (id []primitive.ObjectID) {
	for _, tag := range list {
		if len(tag.Params) < index+1 {
			continue
		}

		if cid := tag.Params[index]; primitive.IsValidObjectID(cid) {
			objID, _ := primitive.ObjectIDFromHex(cid)
			id = append(id, objID)
		}
	}
	return
}
