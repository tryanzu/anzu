package post

import (
	"errors"
	"math"

	"github.com/tryanzu/core/core/common"
	"gopkg.in/mgo.v2/bson"
)

// PostNotFound err.
var PostNotFound = errors.New("post has not been found by given criteria")

func FindId(deps deps, id bson.ObjectId) (post Post, err error) {
	err = deps.Mgo().C("posts").FindId(id).One(&post)
	return
}

func FindList(deps deps, scopes ...common.Scope) (list Posts, err error) {
	err = deps.Mgo().C("posts").Find(common.ByScope(scopes...)).All(&list)
	return
}

func FindRateList(d deps, date string, offset, limit int) ([]bson.ObjectId, error) {
	list := []bson.ObjectId{}
	scores, err := d.LedisDB().ZRangeByScore([]byte("posts:"+date), 0, math.MaxInt64, offset, limit)
	if err != nil {
		return list, err
	}
	for _, n := range scores {
		id := bson.ObjectIdHex(string(n.Member))
		list = append(list, id)
	}
	return list, err
}
