package store

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"time"
)

type Activity struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	LeadId    bson.ObjectId `bson:"lead_id" json:"lead_id"`
	Content   string        `bson:"content" json:"content" binding:"required"`
	Date      time.Time     `bson:"date" json:"date" binding:"required"`
	Completed bool          `bson:"completed" json:"completed"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
	Updated   time.Time     `bson:"updated_at" json:"updated_at"`
}

type Activities []Activity

func (list Activities) ToMap() map[string]Activity {
	m := map[string]Activity{}
	for _, item := range list {
		m[item.Id.Hex()] = item
	}

	return m
}

func (list Activities) QueryLeads() Query {
	id := make([]bson.ObjectId, len(list))
	for i, activity := range list {
		id[i] = activity.LeadId
	}

	return func(col *mgo.Collection) *mgo.Query {
		return col.Find(bson.M{"_id": bson.M{"$in": id}})
	}
}

func AssignActivity(deps Deps, lead Lead, activity Activity) (Activity, error) {
	if activity.Id.Valid() == false {
		activity.Id = bson.NewObjectId()
	}

	activity.LeadId = lead.Id
	activity.Created = time.Now()
	activity.Updated = time.Now()

	_, err := deps.Mgo().C("activities").UpsertId(activity.Id, activity)
	if err != nil {
		return activity, err
	}

	return activity, nil
}

func FindActivities(deps Deps, betweenDates []time.Time, offset, limit int) (Activities, error) {
	var list Activities
	params := bson.M{}
	if len(betweenDates) == 2 {
		params["date"] = bson.M{
			"$gte": betweenDates[0],
			"$lt":  betweenDates[1],
		}
	}

	err := deps.Mgo().C("activities").Find(params).Limit(limit).Skip(offset).All(&list)
	if err != nil {
		return list, err
	}

	return list, nil
}
