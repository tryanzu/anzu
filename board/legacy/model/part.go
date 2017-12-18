package model

import (
	"gopkg.in/mgo.v2/bson"
)

type PartByModel struct {
	Id              bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name            string        `bson:"name" json:"name"`
	Type            string        `bson:"type" json:"type"`
	Slug            string        `bson:"slug" json:"slug"`
	Price           float64       `bson:"price" json:"price"`
	Manufacturer    string        `bson:"manufacturer,omitempty" json:"manufacturer,omitempty"`
	PartNumber      string        `bson:"partnumber,omitempty" json:"partnumber,omitempty"`
	Model           string        `bson:"model,omitempty" json:"model,omitempty"`
	Socket          string        `bson:"supported_sockets,omitempty" json:"supported_sockets,omitempty"`
	LiquidCooled    bool          `bson:"liquid_cooled,omitempty" json:"liquid_cooled,omitempty"`
	BearingType     string        `bson:"bearing_type,omitempty" json:"bearing_type,omitempty"`
	NoiseLevel      string        `bson:"noise_level,omitempty" json:"noise_level,omitempty"`
	FanRpm          string        `bson:"fan_rpm,omitempty" json:"fan_rpm,omitempty"`
	Speed           string        `bson:"speed,omitempty" json:"speed,omitempty"`
	Size            string        `bson:"size,omitempty" json:"size,omitempty"`
	PriceGb         string        `bson:"gb_price,omitempty" json:"gb_price,omitempty"`
	Cas             string        `bson:"cas,omitempty" json:"cas,omitempty"`
	Voltage         string        `bson:"voltage,omitempty" json:"voltage,omitempty"`
	HeatSpreader    bool          `bson:"heat_spreader,omitempty" json:"heat_spreader,omitempty"`
	Ecc             bool          `bson:"ecc,omitempty" json:"ecc,omitempty"`
	Registered      bool          `bson:"registered,omitempty" json:"registered,omitempty"`
	Color           string        `bson:"color,omitempty" json:"color,omitempty"`
	Chipset         string        `bson:"chipset,omitempty" json:"chipset,omitempty"`
	MemorySlots     string        `bson:"memory_slots,omitempty" json:"memory_slots,omitempty"`
	MemoryType      string        `bson:"memory_type,omitempty" json:"memory_type,omitempty"`
	MaxMemory       string        `bson:"memory_max,omitempty" json:"memory_max,omitempty"`
	RaidSupport     bool          `bson:"raid_support,omitempty" json:"raid_support,omitempty"`
	OnboardVideo    bool          `bson:"onboard_video,omitempty" json:"onboard_video,omitempty"`
	Crossfire       bool          `bson:"crossfire_support,omitempty" json:"crossfire_support,omitempty"`
	SliSupport      bool          `bson:"sli_support,omitempty" json:"sli_support,omitempty"`
	SATA            string        `bson:"sata_6gbs" json:"sata_6gbs"`
	OnboardEthernet string        `bson:"onboard_ethernet,omitempty" json:"onboard_ethernet,omitempty"`
	OnboardUsb3     bool          `bson:"onboard_usb_3,omitempty" json:"onboard_usb_3,omitempty"`
	Capacity        string        `bson:"capacity,omitempty" json:"capacity,omitempty"`
	Interface       string        `bson:"interface,omitempty" json:"interface,omitempty"`
	Cache           string        `bson:"cache,omitempty" json:"cache,omitempty"`
	SsdController   string        `bson:"ssd_controller,omitempty" json:"ssd_controller,omitempty"`
	FormFactor      string        `bson:"form_factor,omitempty" json:"form_factor,omitempty"`
	GbPerDollar     string        `bson:"gb_per_dollar,omitempty" json:"gb_per_dollar,omitempty"`
}
