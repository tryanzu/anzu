package components

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Component struct {
	Id           bson.ObjectId       `bson:"_id,omitempty" json:"id"`
	Name         string              `bson:"name" json:"name"`
	FullName     string              `bson:"full_name" json:"full_name"`
	Slug         string              `bson:"slug" json:"slug"`
	Source       string              `bson:"source" json:"source"`
	External     float64             `bson:"external" json:"external"`
	Type         string              `bson:"type" json:"type"`
	Image        string              `bson:"image" json:"image"`
	PartNumber   string              `bson:"part_number" json:"part_number"`
	Manufacturer string              `bson:"manufacturer" json:"manufacturer"`
	Store        ComponentStoreModel `bson:"store,omitempty" json:"store,omitempty"`
}

var ComponentFields bson.M = bson.M{"_id": 1, "name": 1, "store": 1, "full_name": 1, "slug": 1, "source": 1, "external": 1, "type": 1, "image": 1, "part_number": 1, "manufacturer": 1}

type ComponentModel struct {
	Component `bson:",inline"`

	Images       []string            `bson:"images" json:"images"`
	Activated    bool                `bson:"activated" json:"activated"`
	di           *Module
	generic      []byte
}

type ComponentImageModel struct {
	Url      string `bson:"url" json:"url"`
	Path     string `bson:"path" json:"path"`
	Checksum string `bson:"checksum" json:"checksum"`
}

type ComponentStoreModel struct {
	Vendors map[string]ComponentStoreItemModel `bson:"vendors" json:"vendors"`
	Updated time.Time                          `bson:"updated_at" json:"updated_at"`
}

type ComponentStoreItemModel struct {
	Price    float64 `bson:"price" json:"price"`
	Stock    int     `bson:"stock" json:"stock"`
	Priority int     `bson:"priority" json:"priority"`
}

type ComponentHistoricModel struct {
	Id          bson.ObjectId       `bson:"_id,omitempty" json:"id"`
	ComponentId bson.ObjectId       `bson:"component_id" json:"component_id"`
	Store       ComponentStoreModel `bson:"store" json:"store"`
	Created     time.Time           `bson:"created_at" json:"created_at"`
}

type CommentVotesModel struct {
	Id    string `bson:"_id,omitempty" json:"_id,omitempty"`
	Count int    `bson:"count" json:"count"`
}

type ComponentMotherboardModel struct {
	ComponentModel
}

type ComponentCaseModel struct {
	ComponentModel
}

type ComponentMemoryModel struct {
	ComponentModel
}

type ComponentMonitorModel struct {
	ComponentModel
}

type ComponentPowerSupplyModel struct {
	ComponentModel
}

type ComponentVideoCardModel struct {
	ComponentModel
}

type ComponentCpuCoolerModel struct {
	ComponentModel
}

type ComponentCpuModel struct {
	ComponentModel
}

type ComponentStorageModel struct {
	ComponentModel
}

type ComponentTypeCountModel struct {
	Type      string `bson:"_id" json:"type"`
	Count     int    `bson:"count" json:"count"`
}

type AlgoliaComponentModel struct {
	Id        string `json:"objectID"`
	Name      string `json:"name"`
	FullName  string `json:"full_name"`
	Part      string `json:"part_number"`
	Slug      string `json:"slug"`
	Image     string `json:"image"`
	Type      string `json:"type"`
	Activated bool   `json:"activated"`
}
