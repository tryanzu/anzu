package assets

import (
	"time"

	"github.com/tidwall/buntdb"
	"gopkg.in/mgo.v2/bson"
)

type Asset struct {
	ID       bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Original string        `bson:"original" json:"original"`
	Hosted   string        `bson:"hosted" json:"hosted"`
	MD5      string        `bson:"hash" json:"hash"`
	Status   string        `bson:"status" json:"status"`
	Created  time.Time     `bson:"created_at" json:"created_at"`
	Updated  time.Time     `bson:"updated_at" json:"updated_at"`
}

// Assets list.
type Assets []Asset

func (list Assets) UpdateBuntCache(tx *buntdb.Tx) (err error) {
	for _, u := range list {
		url := u.Original
		if len(u.Hosted) > 0 {
			url = u.Hosted
		}
		_, _, err = tx.Set("asset:"+u.ID.Hex()+":url", url, nil)
		if err != nil {
			return
		}
	}
	return
}
