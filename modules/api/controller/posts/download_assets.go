package posts

import (
	"context"
	"github.com/mitchellh/goamz/s3"
	"github.com/tryanzu/core/board/legacy/model"
	"github.com/tryanzu/core/deps"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

func (this API) savePostImages(from string, post_id primitive.ObjectID) error {

	defer this.Errors.Recover()

	// Get the database interface from the DI
	ctx := context.Background()
	database := deps.Container.Mgo()
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
		name = primitive.NewObjectID().Hex()

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

		err = database.Collection("posts").FindOne(ctx, bson.M{"_id": post_id}).Decode(&post)

		if err == nil {

			post_content := post.Content

			// Replace the url on the comment
			if strings.Contains(post_content, from) {

				content := strings.Replace(post_content, from, amazon_url+path, -1)

				// Update the comment
				_, err = database.Collection("posts").UpdateOne(ctx, bson.M{"_id": post_id}, bson.M{"$set": bson.M{"content": content}})
				if err != nil {
					panic(err)
				}
			}

		}
	}

	response.Body.Close()

	return nil
}
