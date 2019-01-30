package assets

import (
	"github.com/tryanzu/core/core/common"
	"gopkg.in/mgo.v2/bson"
)

func FindList(d Deps, scopes ...common.Scope) (list Assets, err error) {
	err = d.Mgo().C("remote_assets").Find(common.ByScope(scopes...)).All(&list)
	return
}

func FindHash(d Deps, hash string) (asset Asset, err error) {
	err = d.Mgo().C("remote_assets").Find(bson.M{
		"hash": hash,
	}).One(&asset)
	return
}

func FindURLs(d Deps, list ...bson.ObjectId) (common.AssetRefsMap, error) {
	hash := common.AssetRefsMap{}
	missing := []bson.ObjectId{}

	// Attempt to fill hashmap using cache layer first.
	for _, id := range list {
		var ref common.AssetRef
		url, err := d.LedisDB().Get([]byte("asset:" + id.Hex() + ":url"))
		if err != nil || len(url) == 0 {
			// Append to list of missing keys
			missing = append(missing, id)
			continue
		}
		ref.URL = string(url)
		// uo = use original
		uo, err := d.LedisDB().Exists([]byte("asset:" + id.Hex() + ":uo"))
		ref.UseOriginal = err == nil && uo > 0
		hash[id] = ref
	}

	if len(missing) == 0 {
		return hash, nil
	}

	assets, err := FindList(d, common.WithinID(missing))
	if err != nil {
		return hash, err
	}

	err = assets.UpdateCache(d)
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
