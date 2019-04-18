package categories

import (
	"time"

	"github.com/tryanzu/core/core/config"
)

var (
	cachedTree Categories
	cachedAt   *time.Time
)

// MakeTree returns categories tree.
func MakeTree(d deps) Categories {
	if cachedAt == nil || cachedAt.Sub(time.Now()) > time.Minute {
		t := time.Now()
		cachedTree = makeTree(d)
		cachedAt = &t
	}
	return cachedTree
}

func makeTree(d deps) (list Categories) {
	cnf := config.C.Copy()
	err := d.Mgo().C("categories").Find(nil).Sort("order").All(&list)
	if err != nil {
		panic(err)
	}
	parent := list[:0]
	child := []Category{}
	for _, c := range list {
		if c.Parent.Valid() {
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
