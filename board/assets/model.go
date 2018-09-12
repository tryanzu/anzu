package assets

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/mitchellh/goamz/s3"

	"github.com/tidwall/buntdb"
	"gopkg.in/mgo.v2/bson"
)

type Asset struct {
	ID       bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Original string        `bson:"original" json:"original"`
	Hosted   string        `bson:"hosted" json:"hosted"`
	DataType string        `bson:"dataType,omitempty" json:"dataType,omitempty"`
	MD5      string        `bson:"hash" json:"hash"`
	Status   string        `bson:"status" json:"status"`
	Created  time.Time     `bson:"created_at" json:"created_at"`
	Updated  time.Time     `bson:"updated_at" json:"updated_at"`
}

// Replace original URL with asset tag.
func (asset Asset) Replace(content string) string {
	tag := "[asset:" + asset.ID.Hex() + "]"
	return strings.Replace(content, asset.Original, tag, -1)
}

func (asset Asset) URL() string {
	if asset.Status == "awaiting" || asset.Status == "remote" || len(asset.Hosted) == 0 {
		return asset.Original
	}

	return asset.Hosted
}

func (asset Asset) useRemote(deps Deps, reason string) (err error) {
	err = deps.Mgo().C("remote_assets").UpdateId(asset.ID, bson.M{
		"$set": bson.M{
			"status":     "remote",
			"updated_at": time.Now(),
			"comments":   reason,
			"dataType":   asset.DataType,
		},
		"$unset": bson.M{
			"hosted": 1,
			"hash":   1,
		},
	})
	return
}

func (asset Asset) useHosted(deps Deps, url string) (err error) {
	err = deps.Mgo().C("remote_assets").UpdateId(asset.ID, bson.M{
		"$set": bson.M{
			"status":     "hosted",
			"updated_at": time.Now(),
			"hosted":     url,
			"hash":       asset.MD5,
			"dataType":   asset.DataType,
		},
	})
	return
}

func (asset Asset) useRepeated(deps Deps, ref Asset) (err error) {
	err = deps.Mgo().C("remote_assets").UpdateId(asset.ID, bson.M{
		"$set": bson.M{
			"status":     "repeated",
			"updated_at": time.Now(),
			"hash":       ref.MD5,
			"hosted":     ref.Hosted,
			"dataType":   ref.DataType,
		},
	})
	return
}

func (asset Asset) Extension() string {
	u, err := url.Parse(asset.Original)
	if err != nil {
		return ""
	}

	return filepath.Ext(u.Path)
}

// Assets list.
type Assets []Asset

// HostRemotes assets into S3 bucket.
func (list Assets) HostRemotes(deps Deps, related string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	baseURL := deps.Config().UString("amazon.url", "")

	for _, ref := range list {
		res, err := client.Get(ref.Original)
		if err != nil {
			ref.useRemote(deps, err.Error())
			continue
		}

		// Read data from remote stream.
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			ref.useRemote(deps, err.Error())
			continue
		}

		hasher := md5.New()
		hasher.Write(data)
		ref.MD5 = hex.EncodeToString(hasher.Sum(nil))
		duplicated, err := FindHash(deps, ref.MD5)
		if err == nil && len(duplicated.Hosted) > 0 {
			ref.useRepeated(deps, duplicated)
			continue
		}

		// Detect the downloaded file type
		ref.DataType = http.DetectContentType(data)
		if ref.DataType[0:5] != "image" {
			ref.useRemote(deps, "Not an image, using original asset ref")
			res.Body.Close()
			continue
		}

		path := related + "/" + ref.ID.Hex() + ref.Extension()
		err = deps.S3().Put(path, data, ref.DataType, s3.ACL("public-read"))
		if err != nil {
			raven.CaptureErrorAndWait(err, map[string]string{
				"assetID": ref.ID.Hex(),
			})
		}

		ref.useHosted(deps, baseURL+path)
		res.Body.Close()
	}
}

func (list Assets) UpdateBuntCache(tx *buntdb.Tx) (err error) {
	for _, u := range list {
		if u.Status == "awaiting" {
			continue
		}
		url := u.Original
		if len(u.Hosted) > 0 {
			url = u.Hosted
		}
		_, _, err = tx.Set("asset:"+u.ID.Hex()+":url", url, nil)
		if err != nil {
			return
		}

		if u.Status == "remote" {
			_, _, err = tx.Set("asset:"+u.ID.Hex()+":uo", "", nil)
			if err != nil {
				return
			}
		}
	}
	return
}
