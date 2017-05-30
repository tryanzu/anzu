package store

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type BLead struct {
	Id      bson.ObjectId            `bson:"_id,omitempty" json:"id,omitempty"`
	Address string                   `bson:"address" json:"-"`
	Answers map[string]interface{}   `bson:"answers" json:"answers"`
	Step    int                      `bson:"step" json:"step"`
	Changes []map[string]interface{} `bson:"changes" json:"-"`
}

func UpsertBLead(deps Deps, lead BLead) (BLead, error) {
	if lead.Id.Valid() == false {
		lead.Id = bson.NewObjectId()
		lead.Changes = []map[string]interface{}{}
	}

	root := bson.M{
		"answers":    lead.Answers,
		"step":       lead.Step,
		"updated_at": time.Now(),
		"address":    lead.Address,
	}

	_, err := deps.Mgo().C("blead").UpsertId(lead.Id, bson.M{
		"$set":  root,
		"$push": bson.M{"changes": root},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	})

	if err != nil {
		return lead, err
	}

	if IsBLeadCompleted(lead) {
		return BLeadToOrder(deps, lead), nil
	}

	return lead, nil
}

func IsBLeadCompleted(lead BLead) bool {
	answers := lead.Answers

	if _, exists := answers["email"]; !exists {
		return false
	}

	if _, exists := answers["names"]; !exists {
		return false
	}

	if _, exists := answers["phone"]; !exists {
		return false
	}

	return true
}

func BLeadToOrder(deps Deps, lead BLead) BLead {
	name := lead.Answers["names"].(string)
	email := lead.Answers["email"].(string)
	phone := lead.Answers["phone"].(string)
	software := lead.Answers["software"].(string)
	content := lead.Answers["comments"].(string)
	budget := lead.Answers["budget"].(float64)
	priority := lead.Answers["priority"].(string)
	location := lead.Answers["location"].(string)
	usage := lead.Answers["usage"].(string)
	icomponents := lead.Answers["components"].([]interface{})
	ichallenges := lead.Answers["challenges"].([]interface{})
	games := ""

	if g, exists := lead.Answers["gameTypes"]; exists {
		games = g.(string)
	}

	if len(software) > 0 {
		content = content + "\n\n Software requerido: " + software
	}

	components := make([]string, len(icomponents))
	for i, item := range icomponents {
		components[i] = item.(string)
	}

	challenges := make([]string, len(ichallenges))
	for i, item := range ichallenges {
		challenges[i] = item.(string)
	}

	order := OrderModel{
		Id: lead.Id,
		User: OrderUserModel{
			Name:  name,
			Email: email,
			Phone: phone,
			Ip:    lead.Address,
		},
		Content:    content,
		Budget:     int(budget),
		Currency:   "MXN",
		State:      location,
		Games:      []string{games},
		Extra:      components,
		Usage:      usage,
		Reference:  "blead",
		BuyDelay:   0,
		Priority:   priority,
		Challenges: challenges,
		Created:    time.Now(),
		Updated:    time.Now(),
	}

	_, err := deps.Mgo().C("orders").UpsertId(lead.Id, order)
	if err != nil {
		panic(err)
	}

	return lead
}
