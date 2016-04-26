package store

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/olebedev/config"
	"github.com/xuyu/goredis"
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

// Find an order by the provided email
func (module *Module) OrderFinder(mail string) (*One, error) {

	var model *OrderModel

	database := module.Mongo.Database

	// Get the order using it's email
	err := database.C("orders").Find(bson.M{"user.email": mail}).Sort("-updated_at").One(&model)

	if err != nil {

		return nil, exceptions.NotFound{"Invalid order follower. Not found."}
	}

	one := &One{data: model, di: module}

	return one, nil
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
