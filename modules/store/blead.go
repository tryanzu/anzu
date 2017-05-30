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
	var usage, name, email, phone, software, content, priority, location, games string
	var components, challenges []string

	defs := func(m map[string]interface{}, name, def string) string {
		if v, exists := m[name]; exists {
			return v.(string)
		}

		return def
	}

	defl := func(m map[string]interface{}, name string) []string {
		if v, exists := m[name]; exists {
			switch t := v.(type) {
			case []interface{}:
				list := []string{}
				for _, str := range t {
					list = append(list, str.(string))
				}

				return list
			}
		}

		return []string{}
	}

	usage = defs(lead.Answers, "usage", "unknown")
	name = defs(lead.Answers, "name", "unknown")
	email = defs(lead.Answers, "email", "unknown")
	phone = defs(lead.Answers, "phone", "unknown")
	software = defs(lead.Answers, "software", "")
	content = defs(lead.Answers, "comments", "")
	priority = defs(lead.Answers, "priority", "unknown")
	location = defs(lead.Answers, "location", "unknown")
	games = defs(lead.Answers, "gameTypes", "unknown")
	components = defl(lead.Answers, "components")
	challenges = defl(lead.Answers, "challenges")

	if len(software) > 0 {
		content = content + "\n\n Software requerido: " + software
	}

	budget := float64(0)
	if v, exists := lead.Answers["budget"]; exists {
		budget = v.(float64)
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
