package search

import (
	"sort"
	"time"

	"github.com/lestrrat-go/ngram"
	"github.com/op/go-logging"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

var log = logging.MustGetLogger("search")
var usersIndex *ngram.Index

const bufferSize = 256

type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Username string        `bson:"username"`
	Seen     time.Time     `bson:"last_seen_at"`
	Score    float64       `bson:"-"`
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
	log.Info("service starting...")
	go func() {
		var user User
		iter := deps.Container.Mgo().C("users").Find(nil).Sort("-last_seen_at").Limit(bufferSize).Iter()
		usersIndex = ngram.NewIndex(1)
		for iter.Next(&user) {
			log.Debugf("indexing %+v", user)
			err := usersIndex.AddItem(user)
			if err != nil {
				log.Error(err)
			}
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
