package store

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/olebedev/config"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"strconv"
	"strings"
	"time"
)

func Boot() *Module {
	module := &Module{}
	return module
}

type Module struct {
	Mongo  *mongo.Service               `inject:""`
	Errors *exceptions.ExceptionsModule `inject:""`
	Redis  *goredis.Redis               `inject:""`
	Mail   *mail.Module                 `inject:""`
	Config *config.Config               `inject:""`
}

func (module *Module) Order(id bson.ObjectId) (*One, error) {

	var model *OrderModel

	database := module.Mongo.Database

	// Get the order using it's id
	err := database.C("orders").FindId(id).One(&model)

	if err != nil {

		return nil, exceptions.NotFound{"Invalid order id. Not found."}
	}

	one := &One{data: model, di: module}
	one.LoadInvoice()

	return one, nil
}

// Find an order by the provided order type
func (module *Module) OrderFinder(order interface{}) (*One, error) {

	var m *OrderModel
	var query *mgo.Query
	db := module.Mongo.Database
	orders := db.C("orders")

	switch order.(type) {
	case bson.ObjectId:
		query = orders.Find(bson.M{"_id": order.(bson.ObjectId)})
	case bson.M:
		query = orders.Find(order.(bson.M))
	case string:
		query = orders.Find(bson.M{"user.email": order.(string)})
	default:
		panic("Unkown argument type")
	}

	err := query.Sort("-updated_at").One(&m)

	if err != nil {
		return nil, exceptions.NotFound{"Could not found order using criteria."}
	}

	one := &One{data: m, di: module}

	return one, nil
}

func (m *Module) TrackEmailOpened(messageId string, trackId bson.ObjectId, seconds int) {

	db := m.Mongo.Database
	c, err := db.C("orders").Find(bson.M{"messages.postmark_id": messageId}).Count()

	if err == nil && c > 0 {

		err := db.C("orders").Update(bson.M{"messages.postmark_id": messageId}, bson.M{"$set": bson.M{"messages.$.opened_at": time.Now(), "messages.$.otrack_id": trackId, "messages.$.read_seconds": seconds}})

		if err != nil {
			panic(err)
		}
	}
}

func (module *Module) CreateOrder(order OrderModel) {

	database := module.Mongo.Database

	// Set the dates
	order.Created = time.Now()
	order.Updated = time.Now()

	err := database.C("orders").Insert(order)

	if err != nil {
		panic(err)
	}
}

func (m *Module) getNextUpCount() int {

	database := m.Mongo.Database
	now := time.Now()
	then := now.Add(-24 * 30 * time.Hour)

	var awaitingCount struct {
		Count int `bson:"count"`
	}

	// Count deals that need our "response"
	err := database.C("orders").Pipe([]bson.M{
		{
			"$match": bson.M{"deleted_at": bson.M{"$exists": false}, "messages": bson.M{"$exists": true}, "created_at": bson.M{"$gte": then}},
		},
		{
			"$project": bson.M{"lastMessage": bson.M{"$arrayElemAt": []interface{}{"$messages", -1}}},
		},
		{
			"$match": bson.M{"lastMessage.type": "inbound"},
		},
		{
			"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}},
		},
	}).One(&awaitingCount)

	if err != nil {
		panic(err)
	}

	return awaitingCount.Count
}

func (m *Module) getOnHoldCount() int {

	database := m.Mongo.Database

	var onholdCount struct {
		Count int `bson:"count"`
	}

	// Count deals that need our "response"
	err := database.C("orders").Pipe([]bson.M{
		{
			"$match": bson.M{"deleted_at": bson.M{"$exists": false}, "messages": bson.M{"$exists": true}},
		},
		{
			"$project": bson.M{"lastMessage": bson.M{"$arrayElemAt": []interface{}{"$messages", -1}}},
		},
		{
			"$match": bson.M{"lastMessage.type": "text"},
		},
		{
			"$group": bson.M{"_id": nil, "count": bson.M{"$sum": 1}},
		},
	}).One(&onholdCount)

	if err != nil {
		panic(err)
	}

	return onholdCount.Count
}

func (m *Module) getNewCount() int {

	database := m.Mongo.Database
	now := time.Now()
	then := now.Add(-24 * 30 * time.Hour)

	count, err := database.C("orders").Find(bson.M{"messages": bson.M{"$exists": false}, "created_at": bson.M{"$gte": then}}).Count()

	if err != nil {
		panic(err)
	}

	return count
}

func (m *Module) getClosedCount() int {

	database := m.Mongo.Database
	count, err := database.C("orders").Find(bson.M{"messages": bson.M{"$exists": true}, "pipeline.step": 5}).Count()

	if err != nil {
		panic(err)
	}

	return count
}

func (module *Module) GetOrdersAggregation() map[string]interface{} {

	aggregation := map[string]interface{}{
		"onHold":   module.getOnHoldCount(),
		"nextUp":   module.getNextUpCount(),
		"brandNew": module.getNewCount(),
		"closed":   module.getClosedCount(),
	}

	return aggregation
}

func (module *Module) GetSortedOrders(limit, skip int, search, group string) []OrderModel {

	var list []OrderModel

	database := module.Mongo.Database
	clause := bson.M{"deleted_at": bson.M{"$exists": false}}

	if group != "" && (group == "brandNew" || group == "nextUp" || group == "onHold") {

		switch group {
		case "brandNew":
			clause["messages"] = bson.M{"$exists": false}
		case "closed":
			clause["pipeline.step"] = 5
		case "nextUp":
			var boundaries []struct {
				Id bson.ObjectId `bson:"_id"`
			}

			err := database.C("orders").Pipe([]bson.M{
				{
					"$match": bson.M{"deleted_at": bson.M{"$exists": false}, "messages": bson.M{"$exists": true}},
				},
				{
					"$sort": bson.M{"updated_at": -1},
				},
				{
					"$project": bson.M{"lastMessage": bson.M{"$arrayElemAt": []interface{}{"$messages", -1}}},
				},
				{
					"$match": bson.M{"lastMessage.type": "inbound"},
				},
				{
					"$skip": skip,
				},
				{
					"$limit": limit,
				},
			}).All(&boundaries)

			if err != nil {
				panic(err)
			}

			var list []bson.ObjectId

			for _, item := range boundaries {
				list = append(list, item.Id)
			}

			clause["_id"] = bson.M{"$in": list}

		case "onHold":
			var boundaries []struct {
				Id bson.ObjectId `bson:"_id"`
			}

			err := database.C("orders").Pipe([]bson.M{
				{
					"$match": bson.M{"deleted_at": bson.M{"$exists": false}, "messages": bson.M{"$exists": true}},
				},
				{
					"$sort": bson.M{"updated_at": -1},
				},
				{
					"$project": bson.M{"lastMessage": bson.M{"$arrayElemAt": []interface{}{"$messages", -1}}},
				},
				{
					"$match": bson.M{"lastMessage.type": "text"},
				},
				{
					"$skip": skip,
				},
				{
					"$limit": limit,
				},
			}).All(&boundaries)

			if err != nil {
				panic(err)
			}

			var list []bson.ObjectId

			for _, item := range boundaries {
				list = append(list, item.Id)
			}

			clause["_id"] = bson.M{"$in": list}
		}
	}

	if search != "" {

		n, err := strconv.Atoi(search)

		if err != nil {
			n = -1
		}

		clause = bson.M{
			"$or": []bson.M{
				{
					"$text": bson.M{
						"$search": search,
					},
				},
				{
					"budget": n,
				},
			},
		}
	}

	query := database.C("orders").Find(clause).Select(bson.M{"score": bson.M{"$meta": "textScore"}}).Sort("$textScore:score", "-updated_at")

	if _, aggregated := clause["_id"]; !aggregated {
		query = query.Limit(limit).Skip(skip)
	}

	err := query.All(&list)
	if err != nil {
		panic(err)
	}

	var mails []string
	var leads []Lead

	for _, order := range list {
		mails = append(mails, order.User.Email)
	}

	err = database.C("leads").Find(bson.M{"email": bson.M{"$in": mails}}).All(&leads)

	if err == nil {

		for index, order := range list {

			for _, lead := range leads {

				if strings.ToLower(lead.Email) == strings.ToLower(order.User.Email) {

					list[index].Lead = true
				}
			}
		}
	}

	return list
}
