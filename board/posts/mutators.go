package post

import (
	"strconv"

	"github.com/siddontang/ledisdb/ledis"
	"github.com/tryanzu/core/board/activity"
	"github.com/tryanzu/core/core/common"
	"gopkg.in/mgo.v2/bson"
)

// TrackView for a post/user.
func TrackView(d deps, id, user bson.ObjectId) (err error) {
	err = activity.Track(d, activity.M{
		RelatedID: id,
		Event:     "post",
		UserID:    user,
	})
	err = syncCountCache(d, "posts:views:"+id.Hex(), bson.M{
		"related_id": id,
		"event":      "post",
	})
	return SyncRates(d, []bson.ObjectId{id})
}

func TrackReachedList(d deps, list []bson.ObjectId, user bson.ObjectId) (err error) {
	err = activity.Track(d, activity.M{
		List:   list,
		Event:  "feed",
		UserID: user,
	})
	for _, r := range list {
		err = syncCountCache(d, "posts:reached:"+r.Hex(), bson.M{
			"list":  r,
			"event": "feed",
		})
		if err != nil {
			return
		}
	}
	return SyncRates(d, list)
}

func SyncRates(d deps, list []bson.ObjectId) error {
	posts, err := FindList(d, common.WithinID(list))
	if err != nil {
		return err
	}
	dates := map[string]struct{}{}
	db := d.LedisDB()
	for _, post := range posts {
		var (
			views   int
			reached int
		)
		id := post.Id.Hex()
		if n, err := db.Get([]byte("posts:views:" + id)); err == nil {
			views, _ = strconv.Atoi(string(n))
		}
		if n, err := db.Get([]byte("posts:reached:" + id)); err == nil {
			reached, _ = strconv.Atoi(string(n))
		}
		viewR := 100.0 / float64(reached) * float64(views)
		date := post.Updated.Format("2006-01-02")
		_, err = db.ZAdd([]byte("posts:"+date), ledis.ScorePair{
			Score:  int64(viewR * 1000000),
			Member: []byte(id),
		})
		if err != nil {
			return err
		}
		dates[date] = struct{}{}
	}

	return err
}

func syncCountCache(d deps, key string, query bson.M) error {
	// Sync post views cache
	v, err := d.LedisDB().Get([]byte(key))
	if err == nil && len(v) > 0 {
		inc, err := strconv.Atoi(string(v))
		if err != nil {
			panic(err)
		}
		inc = inc + 1
		err = d.LedisDB().Set([]byte(key), []byte(strconv.Itoa(inc)))
		return err
	}
	n := activity.Count(d, query)
	return d.LedisDB().Set([]byte(key), []byte(strconv.Itoa(n)))
}
