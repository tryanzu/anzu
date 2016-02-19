package posts

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/mitchellh/goamz/s3"
	"gopkg.in/mgo.v2/bson"

	"errors"
	"net/http"
	"net/url"
	"crypto/tls"
	"path/filepath"
	"fmt"
	"io/ioutil"
	"strings"
)

func (this API) savePostImages(from string, post_id bson.ObjectId) error {

	defer this.Errors.Recover()

	// Get the database interface from the DI
	database := this.Mongo.Database
	amazon_url, err := this.Config.String("amazon.url")

	if err != nil {
		panic(err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Download the file
	response, err := client.Get(from)
	if err != nil {
		return errors.New(fmt.Sprint("Error while downloading", from, "-", err))
	}

	// Read all the bytes to the image
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.New(fmt.Sprint("Error while downloading", from, "-", err))
	}

	// Detect the downloaded file type
	dataType := http.DetectContentType(data)

	if dataType[0:5] == "image" {

		var extension, name string

		// Parse the filename
		u, err := url.Parse(from)

		if err != nil {
			return errors.New(fmt.Sprint("Error while parsing url", from, "-", err))
		}

		extension = filepath.Ext(u.Path)
		name = bson.NewObjectId().Hex()

		if extension != "" {

			name = name + extension
		} else {

			// If no extension is provided on the url then add a dummy one
			name = name + ".jpg"
		}

		path := "posts/" + name
		err = this.S3.Put(path, data, dataType, s3.ACL("public-read"))

		if err != nil {

			panic(err)
		}

		var post model.Post

		err = database.C("posts").Find(bson.M{"_id": post_id}).One(&post)

		if err == nil {

			post_content := post.Content

			// Replace the url on the comment
			if strings.Contains(post_content, from) {

				content := strings.Replace(post_content, from, amazon_url+path, -1)

				// Update the comment
				database.C("posts").Update(bson.M{"_id": post_id}, bson.M{"$set": bson.M{"content": content}})
			}

		}
	}

	response.Body.Close()

	return nil
}