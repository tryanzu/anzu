package categories

import (
	"gopkg.in/mgo.v2/bson"
)

// Category model.
type Category struct {
	ID          bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	Slug        string        `bson:"slug" json:"slug"`
	Color       string        `bson:"color" json:"color"`
	Permissions ACL           `bson:"permissions" json:"-"`
	Parent      bson.ObjectId `bson:"parent,omitempty" json:"parent,omitempty"`
	ReactSet    []string      `bson:"reactSet" json:"-"`
	Order       int           `bson:"order,omitempty" json:"order,omitempty"`

	// Runtime computed properties.
	Child     Categories `bson:"-" json:"subcategories,omitempty"`
	Writable  bool       `bson:"-" json:"writable"`
	Reactions []string   `bson:"-" json:"reactions,omitempty"`
}

// Categories list.
type Categories []Category

// ACL for categories.
type ACL struct {
	Read  []string `bson:"read" json:"read"`
	Write []string `bson:"write" json:"write"`
}

// CheckWrite permissions for categories tree.
func (slice Categories) CheckWrite(fn func([]string) bool) Categories {
	list := make(Categories, len(slice))
	for n, c := range slice {
		list[n] = c
		list[n].Writable = fn(c.Permissions.Write)
		if len(c.Child) > 0 {
			list[n].Child = list[n].Child.CheckWrite(fn)
		}
	}
	return list
}

func (slice Categories) Len() int {
	return len(slice)
}

func (slice Categories) Less(i, j int) bool {
	return slice[i].Order < slice[j].Order
}

func (slice Categories) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
