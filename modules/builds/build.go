package builds

import (
	"gopkg.in/mgo.v2/bson"

	"time"
)

type Build struct {
	Id          bson.ObjectId   `bson:"_id,omitempty" json:"id,omitempty"`
	Ref         string          `bson:"ref" json:"ref"`
	UserId      *bson.ObjectId  `bson:"user_id,omitempty" json:"-"`
	SessionId   string          `bson:"session_id" json:"-"`
	Name        string          `bson:"name" json:"name"`
	Validated   bool            `bson:"validated" json:"validated"`
	Completed   bool            `bson:"completed" json:"completed"`
	CPU         bson.ObjectId   `bson:"cpu,omitempty" json:"-"`
	Cooler      bson.ObjectId   `bson:"cooler,omitempty" json:"-"`
	Motherboard bson.ObjectId   `bson:"motherboard,omitempty" json:"-"`
	GPU         []bson.ObjectId `bson:"gpu,omitempty" json:"-"`
	Memory      []bson.ObjectId `bson:"memory,omitempty" json:"-"`
	Storage     []bson.ObjectId `bson:"storage,omitempty" json:"-"`
	Case        bson.ObjectId   `bson:"case,omitempty" json:"-"`
	PSU         bson.ObjectId   `bson:"psu,omitempty" json:"-"`
	Created     time.Time       `bson:"created_at" json:"created_at"`
	Updated     time.Time       `bson:"updated_at,omitempty" json:"updated_at,omitempty"`

	di *Module
}

func (b *Build) SetDI(m *Module) {
	b.di = m
}

func (b *Build) UpdateByMap(m map[string]interface{}) error {

	set := bson.M{}
	unset := bson.M{}
	db := b.di.Mongo.Database

	if name, exists := m["name"].(string); exists {
		if len(name) > 0 {
			set["name"] = name
		}
	}

	var simpleIdAttrs = []string{"cpu", "cooler", "motherboard", "case", "psu"}

	for _, component := range simpleIdAttrs {
		if c, exists := m[component].(string); exists {
			if bson.IsObjectIdHex(c) {
				set[component] = bson.ObjectIdHex(c)
			}
		}
		if c, exists := m[component].(bool); exists {
			if c == false {
				unset[component] = ""
			}
		}
	}

	var multipleIdAttrs = []string{"gpu", "memory", "storage"}

	for _, component := range multipleIdAttrs {
		if c, exists := m[component].([]string); exists {

			ls := []bson.ObjectId{}

			for _, id := range c {
				if bson.IsObjectIdHex(id) {
					ls = append(ls, bson.ObjectIdHex(id))
				}
			}

			set[component] = ls
		}

		if c, exists := m[component].(bool); exists {
			if c == false {
				unset[component] = ""
			}
		}
	}

	if len(set) > 0 || len(unset) > 0 {

		params := bson.M{}

		if len(set) > 0 {
			params["$set"] = set
		}

		if len(unset) > 0 {
			params["$unset"] = unset
		}

		err := db.C("builds").Update(bson.M{"_id": b.Id}, params)

		if err != nil {
			return err
		}
	}

	return nil
}
