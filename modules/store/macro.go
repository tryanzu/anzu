package store

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

const MACRO_COL = "builds_responses"

type Macro struct {
	Id      bson.ObjectId          `bson:"_id,omitempty" json:"id,omitempty"`
	Title   string                 `bson:"title" json:"title" binding:"required"`
	Content string                 `bson:"content" json:"content" binding:"required"`
	Flags   map[string]interface{} `bson:"flags" json:"flags"`
	Created time.Time              `bson:"created_at" json:"created_at"`
	Updated time.Time              `bson:"updated_at" json:"updated_at"`
}

// Find macros by clause.
func FindMacros(deps Deps, clause bson.M) (macro []Macro, err error) {
	err = deps.Mgo().C(MACRO_COL).Find(clause).All(&macro)
	return
}

// Find macro by clause.
func FindMacro(deps Deps, clause bson.M) (macro Macro, err error) {
	err = deps.Mgo().C(MACRO_COL).Find(clause).One(&macro)
	return
}

// Delete macros by criteria.
func DeleteMacros(deps Deps, clause bson.M) (removed int, err error) {
	changeset, err := deps.Mgo().C(MACRO_COL).RemoveAll(clause)
	removed = changeset.Removed
	return
}

// Update or insert macro as needed.
func UpsertMacro(deps Deps, m Macro) (macro Macro, err error) {
	if m.Id.Valid() == false {
		m.Id = bson.NewObjectId()
		m.Created = time.Now()
	}

	m.Updated = time.Now()
	_, err = deps.Mgo().C(MACRO_COL).UpsertId(m.Id, m)
	macro = m
	return
}
