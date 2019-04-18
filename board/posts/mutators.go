package post

import (
	"log"
	"strconv"
	"time"

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
	err = incrCountCache(d, "posts:views:"+id.Hex(), bson.M{
		"related_id": id,
		"event":      "post",
	})
	if err != nil {
		return
	}
	return SyncRates(d, "views", []bson.ObjectId{id})
}

func TrackReachedList(d deps, list []bson.ObjectId, user bson.ObjectId) (err error) {
	err = activity.Track(d, activity.M{
		List:   list,
		Event:  "feed",
		UserID: user,
	})
	if err != nil {
		return
	}
	for _, r := range list {
		err = incrCountCache(d, "posts:reached:"+r.Hex(), bson.M{
			"list":  r,
			"event": "feed",
		})
		if err != nil {
			return
		}
	}
	return SyncRates(d, "reached", list)
}

func getRelReachedCount(d deps, at time.Time) int64 {
	y, w := at.ISOWeek()
	rateK := []byte("rates:reached:" + strconv.Itoa(y) + "/" + strconv.Itoa(w))
	if v, err := d.LedisDB().Exists(rateK); err == nil && v > 0 {
		kv, err := d.LedisDB().Get(rateK)
		if err != nil {
			panic(err)
		}

		n, err := strconv.Atoi(string(kv))
		if err != nil {
			panic(err)
		}

		return int64(n)
	}
	week := firstDayOfISOWeek(y, w, at.Location())
	endOfWeek := week.AddDate(0, 0, 7)
	n := activity.CountList(d, bson.M{
		"event":      "feed",
		"created_at": bson.M{"$gte": week, "$lte": endOfWeek},
	})
	err := d.LedisDB().Set([]byte(rateK), []byte(strconv.Itoa(n)))
	if err != nil {
		panic(err)
	}
	return int64(n)
}

func getRelViewsCount(d deps, at time.Time) int64 {
	y, w := at.ISOWeek()
	rateK := []byte("rates:views:" + strconv.Itoa(y) + "/" + strconv.Itoa(w))
	if v, err := d.LedisDB().Exists(rateK); err == nil && v > 0 {
		kv, err := d.LedisDB().Get(rateK)
		if err != nil {
			panic(err)
		}

		n, err := strconv.Atoi(string(kv))
		if err != nil {
			panic(err)
		}

		return int64(n)
	}
	week := firstDayOfISOWeek(y, w, at.Location())
	endOfWeek := week.AddDate(0, 0, 7)
	n := activity.Count(d, bson.M{
		"event":      "post",
		"created_at": bson.M{"$gte": week, "$lte": endOfWeek},
	})
	err := d.LedisDB().Set([]byte(rateK), []byte(strconv.Itoa(n)))
	if err != nil {
		panic(err)
	}
	return int64(n)
}

func SyncRates(d deps, kind string, list []bson.ObjectId) error {
	posts, err := FindList(d, common.WithinID(list))
	if err != nil {
		return err
	}
	db := d.LedisDB()
	now := time.Now()
	date := now.Format("2006-01-02")
	y, w := now.ISOWeek()
	relReached := getRelReachedCount(d, now)
	relViews := getRelViewsCount(d, now)
	rateK := []byte("rates:" + kind + ":" + strconv.Itoa(y) + "/" + strconv.Itoa(w))
	v, err := d.LedisDB().Exists(rateK)
	if err == nil && v > 0 && len(list) > 0 {
		count := int64(len(list))
		_, err = d.LedisDB().IncrBy(rateK, count)
		if err != nil {
			panic(err)
		}

		switch kind {
		case "views":
			relViews += count
		case "reached":
			relReached += count
		}
	}
	scores := []ledis.ScorePair{}
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
		if reached == 0 || relReached == 0 || relViews == 0 {
			continue
		}
		// Conversion rates calculation.
		var (
			viewR = 100 / float64(reached) * float64(views)
			//commentR float64
			//usersR   float64
		)
		/*if views > 0 {
			commentR = 100 / float64(views) * float64(post.Comments.Count)
		}
		if post.Comments.Count > 0 {
			usersR = 100 / float64(post.Comments.Count) * float64(len(post.Users))
		}*/

		// Relative conversion rates calculation.
		var (
			reachRR = float64(reached) / float64(relReached)
			viewRR  = float64(views) / float64(relViews)
		)
		rate := ((viewRR * viewR) + reachRR) / 2
		scores = append(scores, ledis.ScorePair{
			Score:  int64(rate * 10000),
			Member: []byte(id),
		})
	}

	log.Printf("[RATES] Saving rates at (rel views: %v reached: %v) %s\n", relViews, relReached, date)
	_, err = db.ZAdd([]byte("posts:"+date), scores...)
	return err
}

func incrCountCache(d deps, key string, query bson.M) error {
	// Sync post views cache
	k := []byte(key)
	v, err := d.LedisDB().Exists(k)
	if err == nil && v > 0 {
		_, err = d.LedisDB().Incr(k)
		return err
	}
	n := activity.Count(d, query)
	return d.LedisDB().Set([]byte(key), []byte(strconv.Itoa(n)))
}

func firstDayOfISOWeek(year int, week int, timezone *time.Location) time.Time {
	date := time.Date(year, 0, 0, 0, 0, 0, 0, timezone)
	isoYear, isoWeek := date.ISOWeek()
	for date.Weekday() != time.Monday { // iterate back to Monday
		date = date.AddDate(0, 0, -1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoYear < year { // iterate forward to the first day of the first week
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoWeek < week { // iterate forward to the first day of the given week
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	return date
}
