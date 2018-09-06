package content

import (
	"log"
	"regexp"
	"strings"

	"github.com/tryanzu/core/board/assets"
	"github.com/tryanzu/core/core/common"
	"gopkg.in/mgo.v2/bson"
)

var (
	assetURL, _ = regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
)

func postReplaceAssetTags(deps Deps, c Parseable, list tags) (processed Parseable, err error) {
	processed = c
	if len(list) == 0 {
		return
	}

	assetList := list.withTag("asset")
	ids := assetList.getIdParams(0)
	if len(ids) == 0 {
		return
	}
	log.Printf("%+v\n", list)
	var urls common.AssetsStringMap
	urls, err = assets.FindURLs(deps, ids...)
	if err != nil {
		return
	}

	content := processed.GetContent()
	for _, tag := range assetList {
		if id := tag.Params[0]; bson.IsObjectIdHex(id) {
			url, exists := urls[bson.ObjectIdHex(id)]
			if exists == false {
				continue
			}

			link := `<img class="asset" src="` + url + `" data-id="` + id + `" alt=""/>`
			content = strings.Replace(content, tag.Original, link, -1)
		}
	}

	processed = processed.UpdateContent(content)
	return
}

/*

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
*/
