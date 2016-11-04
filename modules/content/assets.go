package content

import (
	"github.com/mitchellh/goamz/s3"
	"gopkg.in/mgo.v2/bson"

	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var urlsRegexp, _ = regexp.Compile(`http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`)

type Asset struct {
	Id       bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Original string        `bson:"original" json:"original"`
	Hosted   string        `bson:"hosted" json:"hosted"`
	MD5      string        `bson:"hash" json:"hash"`
	Status   string        `bson:"status" json:"status"`
	Created  time.Time     `bson:"created_at" json:"created_at"`
	Updated  time.Time     `bson:"updated_at" json:"updated_at"`
}

func (self Module) AsyncAssetDownload(o Parseable) bool {

	c := o.GetContent()
	assets := urlsRegexp.FindAllString(c, -1)

	for _, asset := range assets {

		a := self.RegisterOwnAsset(asset, o)
		tag := "[asset:" + a.Id.Hex() + "]"

		c = strings.Replace(c, asset, tag, -1)
	}

	o.UpdateContent(c)
	o.OnParseFilterFinished("assets")

	return true
}

func (self Module) RegisterOwnAsset(remoteUrl string, o Parseable) *Asset {

	asset := &Asset{
		Id:       bson.NewObjectId(),
		Original: remoteUrl,
		Status:   "awaiting",
		Hosted:   "",
		MD5:      "",
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	database := self.Mongo.Database
	err := database.C("remote_assets").Insert(asset)

	if err != nil {
		panic(err)
	}

	// Run process to own asset in bg
	go func(remoteUrl string, module Module, asset *Asset, o Parseable) {

		defer module.Errors.Recover()

		// Get the database interface from the DI
		database := module.Mongo.Database
		amazon_url, err := module.Config.String("amazon.url")

		if err != nil {
			panic(err)
		}

		fail := func(msg error) {

			fmt.Println(msg)

			err := database.C("remote_assets").Update(bson.M{"_id": asset.Id}, bson.M{"$set": bson.M{"status": "remote", "hosted": "", "hash": "", "message": msg.Error()}})

			if err != nil {
				panic(err)
			}
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}

		// Download the file
		response, err := client.Get(remoteUrl)

		if err != nil {
			fail(err)
			return
		}

		// Read all the bytes to the image
		data, err := ioutil.ReadAll(response.Body)

		if err != nil {
			fail(err)
			return
		}

		// Detect the downloaded file type
		dataType := http.DetectContentType(data)

		fmt.Printf("%v is %v \n", asset.Id.Hex(), dataType)

		if dataType[0:5] == "image" {

			var extension, name string

			// Parse the filename
			u, err := url.Parse(remoteUrl)

			if err != nil {
				fail(err)
				return
			}

			extension = filepath.Ext(u.Path)
			name = asset.Id.Hex()

			if extension != "" {
				name = name + extension
			} else {

				// If no extension is provided on the url then add a dummy one
				name = name + ".jpg"
			}

			path := "posts/" + name
			err = module.S3.Put(path, data, dataType, s3.ACL("public-read"))

			if err != nil {
				fail(err)
				panic(err)
			}

			hasher := md5.New()
			hasher.Write(data)

			hash := hex.EncodeToString(hasher.Sum(nil))

			var ra Asset
			err = database.C("remote_assets").Find(bson.M{"hash": hash}).One(&ra)

			fmt.Printf("%v hash is %v and err is %v \n", asset.Id.Hex(), hash, err)

			if err == nil {

				err := database.C("remote_assets").Update(bson.M{"_id": asset.Id}, bson.M{"$set": bson.M{"status": "repeated", "hosted": ra.Hosted, "hash": hash}})

				if err != nil {
					panic(err)
				}

			} else {

				err := database.C("remote_assets").Update(bson.M{"_id": asset.Id}, bson.M{"$set": bson.M{"status": "hosted", "hosted": amazon_url + path, "hash": hash}})

				if err != nil {
					panic(err)
				}
			}
		}

		fail(errors.New("Could not download from remote and self hold"))

		response.Body.Close()

	}(remoteUrl, self, asset, o)

	return asset
}
