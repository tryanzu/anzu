package components

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type ComponentModel struct {
	Id           bson.ObjectId         `bson:"_id,omitempty" json:"id"`
	Name         string                `bson:"name" json:"name"`
	FullName     string                `bson:"full_name" json:"full_name"`
	Slug         string                `bson:"slug" json:"slug"`
	Source       string                `bson:"source" json:"source"`
	Price        float64               `bson:"price" json:"price"`
	External     float64               `bson:"external" json:"external"`
	Type         string                `bson:"type" json:"type"`
	PartNumber   string                `bson:"part_number" json:"part_number"`
	Manufacturer string                `bson:"manufacturer" json:"manufacturer"`
	Images       []ComponentImageModel `bson:"images" json:"images"`
	ImageUrls    []string              `bson:"image_urls" json:"image_urls"`
	Activated    bool `bson:"activated" json:"activated"`
	Store        ComponentStoreModel   `bson:"store,omitempty" json:"store,omitempty"`
	di           *Module
}

type ComponentImageModel struct {
	Url      string `bson:"url" json:"url"`
	Path     string `bson:"path" json:"path"`
	Checksum string `bson:"checksum" json:"checksum"`
}

type ComponentStoreModel struct {
	Prices  map[string]float64 `bson:"prices" json:"prices"`
	Stock   int 			   `bson:"stock" json:"stock"`
	Updated time.Time 		   `bson:"updated_at" json:"updated_at"`
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
