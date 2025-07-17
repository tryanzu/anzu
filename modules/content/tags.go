package content

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/tryanzu/core/deps"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var tag_regex, _ = regexp.Compile(`(?i)\[([a-z0-9]+(:?))+\]`)
var tag_params_regex, _ = regexp.Compile(`(?i)(([a-z0-9]+)(:?))+?`)

type Tag struct {
	Original string
	Name     string
	Params   []string
}

func (self Module) ParseMentionTags(o Parseable, tags []Tag) bool {

	if len(tags) > 0 {

		var ids []string
		c := o.GetContent()

		for _, tag := range tags {

			// Ensure tag for mentions and its params
			if tag.Name == "mention" && len(tag.Params) > 0 {
				if id := tag.Params[0]; len(id) == 24 {
					if _, err := primitive.ObjectIDFromHex(id); err == nil {
						ids = append(ids, id)
					}
				}
			}
		}

		users := self.FetchUsersHelper(ids)

		for _, tag := range tags {

			// Ensure tag for mentions and its params
			if tag.Name == "mention" && len(tag.Params) > 0 {
				if id := tag.Params[0]; len(id) == 24 {
					if _, err := primitive.ObjectIDFromHex(id); err == nil {
						usr, exists := users[id]
						if exists {
							link := `<a class="user-mention" data-id="` + id + `" data-username="` + usr + `">@` + usr + `</a>`
							c = strings.Replace(c, tag.Original, link, -1)
						}
					}
				}
			}
		}

		o.UpdateContent(c)
	}

	return true
}

func (self Module) ParseAssetTags(o Parseable, tags []Tag) bool {

	if len(tags) > 0 {

		var ids []string
		c := o.GetContent()

		for _, tag := range tags {

			// Ensure tag for mentions and its params
			if tag.Name == "asset" && len(tag.Params) > 0 {
				if id := tag.Params[0]; len(id) == 24 {
					if _, err := primitive.ObjectIDFromHex(id); err == nil {
						ids = append(ids, id)
					}
				}
			}
		}

		assets := self.FetchAssetsHelper(ids)

		for _, tag := range tags {

			// Ensure tag for mentions and its params
			if tag.Name == "asset" && len(tag.Params) > 0 {
				if id := tag.Params[0]; len(id) == 24 {
					if _, err := primitive.ObjectIDFromHex(id); err == nil {
						asset, exists := assets[id]
						if exists {
							link := asset
							c = strings.Replace(c, tag.Original, link, -1)
						}
					}
				}
			}
		}

		o.UpdateContent(c)
	}

	return true
}

func (self Module) FetchUsersHelper(ls []string) map[string]string {

	usrMap := map[string]string{}
	redis := self.Redis
	users, err := redis.HMGet("usernamesHash", ls...)

	if err == nil && len(users) > 0 {

		for index, id := range ls {

			// ls index should match users index
			usr := users[index]

			if len(usr) > 0 {
				usrMap[id] = string(usr)
			}
		}
	}

	missing := []primitive.ObjectID{}

	for _, id := range ls {
		if _, exists := usrMap[id]; !exists {
			if oid, err := primitive.ObjectIDFromHex(id); err == nil {
				missing = append(missing, oid)
			}
		}
	}

	if len(missing) > 0 {
		var targets []struct {
			Id       primitive.ObjectID `bson:"_id"`
			Username string             `bson:"username"`
		}

		ctx := context.Background()
		database := deps.Container.Mgo()
		collection := database.Collection("users")
		filter := bson.M{"_id": bson.M{"$in": missing}}
		opts := options.Find().SetProjection(bson.M{"username": 1})
		cursor, err := collection.Find(ctx, filter, opts)
		if err != nil {
			return usrMap
		}
		defer cursor.Close(ctx)
		err = cursor.All(ctx, &targets)

		if err == nil {
			for _, usr := range targets {
				usrMap[usr.Id.Hex()] = usr.Username

				go func() { _, _ = redis.HSet("usernamesHash", usr.Id.Hex(), usr.Username) }()
			}
		}
	}

	return usrMap
}

func (self Module) FetchAssetsHelper(ls []string) map[string]string {

	assetMap := map[string]string{}
	redis := self.Redis
	assets, err := redis.HMGet("assetsHash", ls...)

	if err == nil && len(assets) > 0 {

		for index, id := range ls {

			// ls index should match users index
			asset := assets[index]

			if len(asset) > 0 {
				assetMap[id] = string(asset)
			}
		}
	}

	missing := []primitive.ObjectID{}

	for _, id := range ls {
		if _, exists := assetMap[id]; !exists {
			if oid, err := primitive.ObjectIDFromHex(id); err == nil {
				missing = append(missing, oid)
			}
		}
	}

	if len(missing) > 0 {
		var targets []Asset

		ctx := context.Background()
		database := deps.Container.Mgo()
		collection := database.Collection("remote_assets")
		filter := bson.M{"_id": bson.M{"$in": missing}}
		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			return assetMap
		}
		defer cursor.Close(ctx)
		err = cursor.All(ctx, &targets)

		if err == nil {
			for _, asset := range targets {

				var url string

				if len(asset.Hosted) > 0 {
					url = asset.Hosted
				} else if len(asset.Original) > 0 {
					url = asset.Original
				}

				assetMap[asset.Id.Hex()] = url

				if asset.Status != "awaiting" {
					go func() { _, _ = redis.HSet("assetsHash", asset.Id.Hex(), url) }()
				}
			}
		}
	}

	return assetMap
}

func (self Module) ParseTags(o Parseable) error {

	start := time.Now()
	c := o.GetContent()
	tags := []Tag{}
	mtags := tag_regex.FindAllString(c, -1)

	for _, tag := range mtags {
		params := tag_params_regex.FindAllString(tag, -1)
		count := len(params) - 1

		for i, param := range params {
			if i != count {
				params[i] = param[:len(param)-1]
			}
		}

		// Check length of params JIC
		if len(params) > 0 {
			tags = append(tags, Tag{
				Original: tag,
				Name:     params[0],
				Params:   params[1:],
			})
		}
	}

	chain := []func(Parseable, []Tag) bool{
		self.ParseMentionTags,
		self.ParseAssetTags,
	}

	for _, fn := range chain {
		next := fn(o, tags)

		if !next {
			break
		}
	}

	o.OnParseFinished()
	elapsed := time.Since(start)

	fmt.Printf("Parse tags took: %v nanoseconds\n", elapsed)

	return nil
}
