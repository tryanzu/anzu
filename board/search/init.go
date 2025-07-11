package search

import (
	"context"
	"sort"
	"time"

	"github.com/lestrrat-go/ngram"
	"github.com/op/go-logging"
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/deps"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var log = logging.MustGetLogger("search")
var usersIndex *ngram.Index

const bufferSize = 256

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Username string             `bson:"username"`
	Seen     time.Time          `bson:"last_seen_at"`
	Score    float64            `bson:"-"`
}

func (u User) Id() string {
	return u.ID.Hex()
}

func (u User) Content() string {
	return u.Username
}

type users []User

func (a users) Len() int           { return len(a) }
func (a users) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a users) Less(i, j int) bool { return a[i].Score > a[j].Score }

func prepare() {
	log.SetBackend(config.LoggingBackend)
	log.Info("service starting...")
	go func() {
		for {
			opts := options.Find().SetSort(bson.M{"last_seen_at": -1}).SetLimit(bufferSize)
			cursor, err := deps.Container.Mgo().Collection("users").Find(context.Background(), bson.M{}, opts)
			if err != nil {
				log.Error(err)
				time.Sleep(10 * time.Minute)
				continue
			}
			usersIndex = ngram.NewIndex(1)
			for cursor.Next(context.Background()) {
				var user User
				err := cursor.Decode(&user)
				if err != nil {
					log.Error(err)
					continue
				}
				err = usersIndex.AddItem(user)
				if err != nil {
					log.Error(err)
				}
			}
			cursor.Close(context.Background())
			log.Info("in-memory user search index has been rehydrated")
			time.Sleep(10 * time.Minute)
		}
	}()
	go func() {
		for {
			<-config.C.Reload
			log.SetBackend(config.LoggingBackend)
		}
	}()
}

func Users(match string) users {
	results := usersIndex.IterateSimilar(match, 0.5, bufferSize)
	list := users{}
	for res := range results {
		u := res.Item.(User)
		u.Score = res.Score
		list = append(list, u)
	}
	sort.Sort(list)
	return list
}

func Boot() {
	prepare()
}
