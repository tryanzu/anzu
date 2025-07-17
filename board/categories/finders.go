package categories

import (
	"context"
	"time"

	"github.com/tryanzu/core/core/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	cachedTree Categories
	cachedAt   *time.Time
)

// MakeTree returns categories tree.
func MakeTree(d deps) Categories {
	if cachedAt == nil || time.Until(*cachedAt) > time.Minute {
		t := time.Now()
		cachedTree = makeTree(d)
		cachedAt = &t
	}
	return cachedTree
}

func makeTree(d deps) (list Categories) {
	cnf := config.C.Copy()
	ctx := context.TODO()
	opt := options.Find().SetSort(bson.M{"order": 1})
	cursor, err := d.Mgo().Collection("categories").Find(ctx, bson.M{}, opt)
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &list)
	if err != nil {
		panic(err)
	}
	parent := list[:0]
	child := []Category{}
	for _, c := range list {
		if !c.Parent.IsZero() {
			c.Reactions = cnf.Site.MakeReactions(c.ReactSet)
			child = append(child, c)
		} else {
			parent = append(parent, c)
		}
	}
	for n, p := range parent {
		matches := []Category{}
		for _, c := range child {
			if c.Parent == p.ID {
				matches = append(matches, c)
			}
		}
		parent[n].Child = matches
	}
	list = parent
	return
}
