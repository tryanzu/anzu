package handle

import (
	"github.com/fernandez14/spartangeek-blacker/model"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"time"
)

type CollectorAPI struct {
	DataService *mongo.Service `inject:""`
}

func (di *CollectorAPI) Activity(thread model.Activity) {

	// Get the database interface from the DI
	database := di.DataService.Database

	// Set the created at field
	thread.Created = time.Now()

	err := database.C("activity").Insert(thread)

	if err != nil {
		panic(err)
	}
}
