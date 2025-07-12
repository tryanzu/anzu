package assets

import (
	"context"
	"time"

	"github.com/tryanzu/core/core/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func FindList(d Deps, scopes ...common.Scope) (list Assets, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := d.Mgo().Collection("remote_assets").Find(ctx, common.ByScope(scopes...))
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &list)
	return
}

func FindHash(d Deps, hash string) (asset Asset, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = d.Mgo().Collection("remote_assets").FindOne(ctx, bson.M{
		"hash": hash,
	}).Decode(&asset)
	if err == mongo.ErrNoDocuments {
		err = nil // Convert to match original behavior
	}
	return
}

func FindURLs(d Deps, list ...primitive.ObjectID) (common.AssetRefsMap, error) {
	hash := common.AssetRefsMap{}
	missing := []primitive.ObjectID{}

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
