package post

import (
	"fmt"
	"strconv"

	"github.com/tidwall/buntdb"
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
	err = d.BuntDB().Update(syncCountCache(d, "posts:views:"+id.Hex(), bson.M{
		"related_id": id,
		"event":      "post",
	}))
	return SyncRates(d, []bson.ObjectId{id})
}

func TrackReachedList(d deps, list []bson.ObjectId, user bson.ObjectId) (err error) {
	err = activity.Track(d, activity.M{
		List:   list,
		Event:  "feed",
		UserID: user,
	})
	for _, r := range list {
		err = d.BuntDB().Update(syncCountCache(d, "posts:reached:"+r.Hex(), bson.M{
			"list":  r,
			"event": "feed",
		}))
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
	err = d.BuntDB().Update(func(tx *buntdb.Tx) error {
		dates := map[string]struct{}{}
		for _, post := range posts {
			var (
				views   int
				reached int
			)
			id := post.Id.Hex()
			if n, err := tx.Get("posts:views:" + id); err == nil {
				views, _ = strconv.Atoi(n)
			}
			if n, err := tx.Get("posts:reached:" + id); err == nil {
				reached, _ = strconv.Atoi(n)
			}
			viewR := 100.0 / float64(reached) * float64(views)
			date := post.Updated.Format("2006-01-02")
			_, _, err = tx.Set("posts:"+date+":"+id, fmt.Sprintf("%.6f", viewR), nil)
			if err != nil {
				return err
			}
			dates[date] = struct{}{}
		}
		for d := range dates {
			tx.CreateIndex("posts:"+d, "posts:"+d+":*", buntdb.IndexFloat)
		}
		return nil
	})
	return err
}

func syncCountCache(d deps, key string, query bson.M) func(*buntdb.Tx) error {
	return func(tx *buntdb.Tx) error {
		// Sync post views cache
		v, err := tx.Get(key)
		if err == nil {
			inc, err := strconv.Atoi(v)
			if err != nil {
				panic(err)
			}
			inc = inc + 1
			_, _, err = tx.Set(key, strconv.Itoa(inc), nil)
			return err
		}
		n := activity.Count(d, query)
		_, _, err = tx.Set(key, strconv.Itoa(n), nil)
		return nil
	}
}
