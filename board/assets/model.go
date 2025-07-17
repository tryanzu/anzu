package assets

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/getsentry/raven-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Asset struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Original string             `bson:"original" json:"original"`
	Hosted   string             `bson:"hosted" json:"hosted"`
	DataType string             `bson:"dataType,omitempty" json:"dataType,omitempty"`
	MD5      string             `bson:"hash" json:"hash"`
	Status   string             `bson:"status" json:"status"`
	Created  time.Time          `bson:"created_at" json:"created_at"`
	Updated  time.Time          `bson:"updated_at" json:"updated_at"`
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = deps.Mgo().Collection("remote_assets").UpdateOne(ctx, bson.M{"_id": asset.ID}, bson.M{
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = deps.Mgo().Collection("remote_assets").UpdateOne(ctx, bson.M{"_id": asset.ID}, bson.M{
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = deps.Mgo().Collection("remote_assets").UpdateOne(ctx, bson.M{"_id": asset.ID}, bson.M{
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
	// TODO: complement with env var
	baseURL := ""

	for _, ref := range list {
		res, err := client.Get(ref.Original)
		if err != nil {
			_ = ref.useRemote(deps, err.Error())
			continue
		}

		// Read data from remote stream.
		data, err := io.ReadAll(res.Body)
		if err != nil {
			_ = ref.useRemote(deps, err.Error())
			continue
		}

		hasher := md5.New()
		hasher.Write(data)
		ref.MD5 = hex.EncodeToString(hasher.Sum(nil))
		duplicated, err := FindHash(deps, ref.MD5)
		if err == nil && len(duplicated.Hosted) > 0 {
			_ = ref.useRepeated(deps, duplicated)
			continue
		}

		// Detect the downloaded file type
		ref.DataType = http.DetectContentType(data)
		if ref.DataType[0:5] != "image" {
			_ = ref.useRemote(deps, "Not an image, using original asset ref")
			res.Body.Close()
			continue
		}

		path := related + "/" + ref.ID.Hex() + ref.Extension()
		err = deps.S3().PutObject(path, data, ref.DataType)
		if err != nil {
			raven.CaptureErrorAndWait(err, map[string]string{
				"assetID": ref.ID.Hex(),
			})
		}

		_ = ref.useHosted(deps, baseURL+path)
		res.Body.Close()
	}
}

func (list Assets) UpdateCache(d Deps) (err error) {
	for _, u := range list {
		if u.Status == "awaiting" {
			continue
		}
		url := u.Original
		if len(u.Hosted) > 0 {
			url = u.Hosted
		}
		err = d.LedisDB().Set([]byte("asset:"+u.ID.Hex()+":url"), []byte(url))
		if err != nil {
			return
		}

		if u.Status == "remote" {
			err = d.LedisDB().Set([]byte("asset:"+u.ID.Hex()+":uo"), []byte(""))
			if err != nil {
				return
			}
		}
	}
	return
}
