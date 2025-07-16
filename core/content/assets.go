package content

import (
	"regexp"
	"strings"

	"github.com/tryanzu/core/board/assets"
	"github.com/tryanzu/core/core/common"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	assetURL, _ = regexp.Compile(`(?m)^http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)
)

func preReplaceAssetTags(d DepsInterface, c Parseable) (processed Parseable, err error) {
	processed = c
	content := processed.GetContent()
	list := assetURL.FindAllString(content, -1)
	if len(list) == 0 {
		return
	}
	tags := assets.Assets{}
	for _, url := range list {
		var asset assets.Asset
		asset, err = assets.FromURL(d, url)
		if err != nil {
			return
		}
		content = asset.Replace(content)
		tags = append(tags, asset)
	}

	// Attempt to host remote assets using S3 in another process
	go tags.HostRemotes(d, "post")
	processed = processed.UpdateContent(content)
	return
}

func postReplaceAssetTags(d DepsInterface, c Parseable, list tags) (processed Parseable, err error) {
	processed = c
	if len(list) == 0 {
		return
	}

	assetList := list.withTag("asset")
	ids := assetList.getIdParams(0)
	if len(ids) == 0 {
		return
	}
	var urls common.AssetRefsMap
	urls, err = assets.FindURLs(d, ids...)
	if err != nil {
		return
	}

	content := processed.GetContent()
	for _, tag := range assetList {
		if id := tag.Params[0]; primitive.IsValidObjectID(id) {
			objID, _ := primitive.ObjectIDFromHex(id)
			ref, exists := urls[objID]
			if exists == false {
				continue
			}
			content = strings.Replace(content, tag.Original, ref.URL, -1)
		}
	}

	processed = processed.UpdateContent(content)
	return
}

/*

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
*/
