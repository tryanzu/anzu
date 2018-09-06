package assets

import (
	"github.com/tidwall/buntdb"
	"github.com/tryanzu/core/core/common"
	"gopkg.in/mgo.v2/bson"
)

func FindList(deps Deps, scopes ...common.Scope) (list Assets, err error) {
	err = deps.Mgo().C("remote_assets").Find(common.ByScope(scopes...)).All(&list)
	return
}

func FindURLs(deps Deps, list ...bson.ObjectId) (common.AssetsStringMap, error) {
	hash := common.AssetsStringMap{}
	missing := []bson.ObjectId{}

	// Attempt to fill hashmap using cache layer first.
	deps.BuntDB().View(func(tx *buntdb.Tx) error {
		for _, id := range list {
			v, err := tx.Get("asset:" + id.Hex() + ":url")
			if err == nil {
				hash[id] = v
				continue
			}

			// Append to list of missing keys
			missing = append(missing, id)
		}

		return nil
	})

	if len(missing) == 0 {
		return hash, nil
	}

	assets, err := FindList(deps, common.WithinID(missing))
	if err != nil {
		return hash, err
	}

	err = deps.BuntDB().Update(assets.UpdateBuntCache)
	if err != nil {
		return hash, err
	}

	for _, u := range assets {
		url := u.Original
		if len(u.Hosted) > 0 {
			url = u.Hosted
		}
		hash[u.ID] = url
	}

	return hash, nil
}
