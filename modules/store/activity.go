package store

import (
	"gopkg.in/mgo.v2/bson"

	"time"
)

type Activity struct {
	Id          bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	LeadId      bson.ObjectId `bson:"lead_id" json:"lead_id"`
	TypeId      string        `bson:"type_id" json:"type_id" binding:"required"`
	TypeName    string        `bson:"type_name" json:"type_name" binding:"required"`
	AssignedTo  string        `bson:"assigned_to" json:"assigned_to"`
	AssignedId  string        `bson:"assigned_id" json:"assigned_id"`
	Notes       string        `bson:"notes" json:"notes"`
	Description string        `bson:"description" json:"description" binding:"required"`
	Date        time.Time     `bson:"date" json:"date" binding:"required"`
	Duration    int           `bson:"duration" json:"duration"`
	Completed   bool          `bson:"completed" json:"completed"`
	Created     time.Time     `bson:"created_at" json:"created_at"`
	Updated     time.Time     `bson:"updated_at" json:"updated_at"`
}

func AssignActivity(deps Deps, lead Lead, activity Activity) (Activity, error) {
	activity.Id = bson.NewObjectId()
	activity.LeadId = lead.Id
	activity.Created = time.Now()
	activity.Updated = time.Now()

	err := deps.Mgo().C("activities").Insert(activity)
	if err != nil {
		return activity, err
	}

	return activity, nil
}

func FindActivities(deps Deps, typeId string, betweenDates []time.Time, offset, limit int) ([]Activity, error) {
	var activities []Activity
	params := bson.M{}
	if len(betweenDates) == 2 {
		params["date"] = bson.M{
			"$gte": betweenDates[0],
			"$lt":  betweenDates[1],
		}
	}

	if len(typeId) > 0 {
		params["type_id"] = typeId
	}

	err := deps.Mgo().C("activities").Find(params).Limit(limit).Skip(offset).All(&activities)
	if err != nil {
		return activities, err
	}

	return activities, nil
}
