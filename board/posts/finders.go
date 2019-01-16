package post

import (
	"errors"
	"strings"

	"github.com/tidwall/buntdb"
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
	d.BuntDB().CreateIndex("posts:"+date, "posts:"+date+":*", buntdb.IndexFloat)
	err := d.BuntDB().View(func(tx *buntdb.Tx) error {
		cursor := 0
		return tx.Ascend("posts:"+date, func(k, v string) bool {
			if cursor < offset {
				cursor++
				return true
			}
			s := strings.Split(k, ":")
			// Invalid key
			if len(s) < 2 || bson.IsObjectIdHex(s[2]) == false {
				return true
			}
			id := bson.ObjectIdHex(s[2])
			list = append(list, id)
			cursor++
			return cursor <= offset+limit
		})
	})
	return list, err
}
