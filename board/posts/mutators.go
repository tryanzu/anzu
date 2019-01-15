package post

import (
	"strconv"

	"github.com/tidwall/buntdb"
	"github.com/tryanzu/core/board/activity"
	"gopkg.in/mgo.v2/bson"
)

// TrackView for a post/user.
func TrackView(d deps, id, user bson.ObjectId) (err error) {
	err = activity.Track(d, activity.M{
		RelatedID: id,
		Event:     "post",
		UserID:    user,
	})
	err = d.BuntDB().Update(func(tx *buntdb.Tx) error {
		k := "posts:views:" + id.Hex()
		v, err := tx.Get(k)
		if err == nil {
			inc, err := strconv.Atoi(v)
			if err != nil {
				panic(err)
			}
			inc = inc + 1
			_, _, err = tx.Set(k, strconv.Itoa(inc), nil)
			return err
		}
		n := activity.Count(d, bson.M{
			"related_id": id,
			"event":      "post",
		})
		_, _, err = tx.Set(k, strconv.Itoa(n), nil)
		return nil
	})
	return
}
