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

	var r mail.InboundMail

	db := m.Mongo.Database
	err := db.C("inbound_mails").Find(bson.M{"messageid": messageId}).One(&r)

	if err == nil {

		err := db.C("orders").Update(bson.M{"messages.related_id": r.Id}, bson.M{"$set": bson.M{"messages.$.opened_at": time.Now(), "messages.$.otrack_id": trackId, "messages.$.read_seconds": seconds}})

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

func (module *Module) GetSortedOrders(limit, skip int, search string) []OrderModel {

	var list []OrderModel

	database := module.Mongo.Database

	clause := bson.M{"deleted_at": bson.M{"$exists": false}}

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

	err := database.C("orders").Find(clause).Select(bson.M{"score": bson.M{"$meta": "textScore"}}).Sort("$textScore:score", "-updated_at").Limit(limit).Skip(skip).All(&list)

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
