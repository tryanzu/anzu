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
	CPU         bson.ObjectId   `bson:"cpu,omitempty" json:"cpu"`
	Cooler      bson.ObjectId   `bson:"cooler,omitempty" json:"cooler"`
	Motherboard bson.ObjectId   `bson:"motherboard,omitempty" json:"motherboard"`
	GPU         []bson.ObjectId `bson:"gpu,omitempty" json:"gpu"`
	Memory      []bson.ObjectId `bson:"memory,omitempty" json:"memory"`
	Storage     []bson.ObjectId `bson:"storage,omitempty" json:"storage"`
	Case        bson.ObjectId   `bson:"case,omitempty" json:"case"`
	PSU         bson.ObjectId   `bson:"psu,omitempty" json:"psu"`
	Created     time.Time       `bson:"created_at" json:"created_at"`
	Updated     time.Time       `bson:"updated_at,omitempty" json:"updated_at,omitempty"`

	Components map[string]interface{} `bson:"-" json:"components,omitempty"`
	di         *Module
}

func (b *Build) SetDI(m *Module) {
	b.di = m
}

func (b *Build) LoadComponents() {

	var components []struct {
		Id       bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
		Name     string        `bson:"name" json:"name"`
		FullName string        `bson:"full_name" json:"full_name"`
		Type     string        `bson:"type" json:"type"`
	}

	var ids []bson.ObjectId

	if b.CPU.Valid() {
		ids = append(ids, b.CPU)
	}
	if b.Cooler.Valid() {
		ids = append(ids, b.Cooler)
	}
	if b.Motherboard.Valid() {
		ids = append(ids, b.Motherboard)
	}
	if b.Case.Valid() {
		ids = append(ids, b.Case)
	}
	if b.PSU.Valid() {
		ids = append(ids, b.PSU)
	}

	for _, id := range b.GPU {
		if id.Valid() {
			ids = append(ids, id)
		}
	}

	for _, id := range b.Memory {
		if id.Valid() {
			ids = append(ids, id)
		}
	}

	for _, id := range b.Storage {
		if id.Valid() {
			ids = append(ids, id)
		}
	}

	db := b.di.Mongo.Database
	err := db.C("components").Find(bson.M{"_id": bson.M{"$in": ids}}).Select(bson.M{"name": 1, "full_name": 1, "type": 1}).All(&components)

	if err != nil {
		panic(err)
	}

	list := make(map[string]interface{}, len(components))

	for _, c := range components {
		list[c.Id.Hex()] = c
	}

	b.Components = list
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

			b.di.Logger.Debugf("%s its inside payload.", component)

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
		if c, exists := m[component].([]interface{}); exists {

			b.di.Logger.Debugf("%s its inside payload.", component)
			ls := []bson.ObjectId{}

			for _, id := range c {
				if bson.IsObjectIdHex(id.(string)) {
					ls = append(ls, bson.ObjectIdHex(id.(string)))
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
