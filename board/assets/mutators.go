package assets

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// FromURL asset.
func FromURL(deps Deps, url string) (ref Asset, err error) {
	ref = Asset{
		ID:       bson.NewObjectId(),
		Original: url,
		Status:   "awaiting",
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	err = deps.Mgo().C("remote_assets").Insert(&ref)
	return
}
