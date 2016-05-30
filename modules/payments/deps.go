package payments

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"gopkg.in/mgo.v2/bson"
)

func GetModule(config map[string]Gateway) *Module {
	return &Module{Gateways: config}
}

type Module struct {
	Mongo    *mongo.Service `inject:""`
	Gateways map[string]Gateway
}

func (m *Module) Get(payment interface{}) (*Payment, error) {

	switch payment.(type) {
	case bson.ObjectId:

		var this *Payment
		database := m.Mongo.Database
		err := database.C("payments").FindId(payment.(bson.ObjectId)).One(&this)

		if err != nil {
			return nil, exceptions.NotFound{"Invalid payment id. Not found."}
		}

		this.SetDI(m)

		return this, nil

	case bson.M:

		var this *Payment
		database := m.Mongo.Database
		err := database.C("posts").Find(payment.(bson.M)).One(&this)

		if err != nil {
			return nil, exceptions.NotFound{"Invalid payment id. Not found."}
		}

		this.SetDI(m)

		return this, nil

	case *Payment:

		this := payment.(*Payment)
		this.SetDI(m)

		return this, nil

	default:
		panic("Unkown argument")
	}
}
