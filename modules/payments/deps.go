package payments

import (
	"github.com/fernandez14/spartangeek-blacker/mongo"
)

type Module struct {
	Mongo *mongo.Service `inject:""`
}
