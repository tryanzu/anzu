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

func FindHash(deps Deps, hash string) (asset Asset, err error) {
	err = deps.Mgo().C("remote_assets").Find(bson.M{
		"hash": hash,
	}).One(&asset)
	return
}

func FindURLs(deps Deps, list ...bson.ObjectId) (common.AssetRefsMap, error) {
	hash := common.AssetRefsMap{}
	missing := []bson.ObjectId{}

	// Attempt to fill hashmap using cache layer first.
	deps.BuntDB().View(func(tx *buntdb.Tx) error {
		for _, id := range list {
			var ref common.AssetRef
			url, err := tx.Get("asset:" + id.Hex() + ":url")
			if err != nil {
				// Append to list of missing keys
				missing = append(missing, id)
				continue
			}
			ref.URL = url
			// uo = use original
			_, err = tx.Get("asset:" + id.Hex() + ":uo")
			ref.UseOriginal = err == nil
			hash[id] = ref
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
		hash[u.ID] = common.AssetRef{
			URL:         u.URL(),
			UseOriginal: u.Status == "remote",
		}
	}

	return hash, nil
}
